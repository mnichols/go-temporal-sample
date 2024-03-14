package orchestrations

import (
	"errors"
	"fmt"
	"github.com/mnichols/go-temporal-sample/pkg/auth"
	"github.com/mnichols/go-temporal-sample/pkg/cicd"
	"github.com/mnichols/go-temporal-sample/pkg/messaging"
	"github.com/mnichols/go-temporal-sample/pkg/notifications"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func (o *Orchestrations) OnboardApplication(ctx workflow.Context, params *messaging.OnboardApp) error {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 3,
	})

	logger := workflow.GetLogger(ctx)

	var getSecretsResponse *messaging.GetSecretsResponse

	var jfrogRequest *messaging.SetupJFrogRequest
	var jfrogResponse *messaging.SetupJFrogResponse

	if err := workflow.ExecuteActivity(ctx, auth.TypeHandlers.GetVaultSecrets, &messaging.GetSecretsRequest{
		UserID: params.UserID,
	}).Get(ctx, &getSecretsResponse); err != nil {
		return fmt.Errorf("failed to get secrets %w", err)
	}

	jfrogRequest = &messaging.SetupJFrogRequest{
		SubscriptionID: params.SubscriptionID,
		Secret:         getSecretsResponse.Secrets["jfrog"],
	}
	var err error
	if jfrogResponse, err = o.setupJFrog(
		ctx,
		getSecretsResponse,
		params,
		jfrogRequest,
		1,
	); err != nil {
		return fmt.Errorf("could not setup jfrog %w", err)
	}

	logger.Debug("jfrog setup %v", jfrogResponse)
	return nil
}

func (o *Orchestrations) setupJFrog(ctx workflow.Context,
	secrets *messaging.GetSecretsResponse,
	params *messaging.OnboardApp,
	request *messaging.SetupJFrogRequest,
	counter int,
) (*messaging.SetupJFrogResponse, error) {

	logger := workflow.GetLogger(ctx)
	if counter > 10 {
		return nil, fmt.Errorf("too many attempts were made trying to setup jfrog %d", counter)
	}

	var response *messaging.SetupJFrogResponse
	var err error
	err = workflow.ExecuteActivity(ctx, cicd.TypeHandlers.SetupJFrog, request).Get(ctx, &response)
	if err == nil {
		// early return we are OK!
		return response, nil
	}

	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		switch appErr.Type() {
		case cicd.BadSubscriptionIDErr:
			chanCtx, cancelCorrection := workflow.WithCancel(ctx)
			correctionSignalChan := workflow.GetSignalChannel(chanCtx, SignalCorrectSubscriptionID)

			var correction *messaging.CorrectSubscriptionIDCommand
			// setup our signal handler
			workflow.Go(ctx, func(ctx workflow.Context) {
				correctionSignalChan.Receive(ctx, &correction)
				logger.Debug("received signal", "signal", SignalCorrectSubscriptionID)
			})
			workflow.Go(ctx, func(ctx workflow.Context) {
				if err := workflow.ExecuteActivity(ctx, notifications.TypeHandlers.RequestCorrection, &messaging.RequestCorrectionRequest{
					BadSubscriptionID: request.SubscriptionID,
					UserID:            params.UserID,
				}); err != nil {
					logger.Error("failed to notify of correction! what should we do???")
				}
			})

			// this timeout can be as long as your secrets would be valid
			ok, err := workflow.AwaitWithTimeout(ctx, time.Minute*60, func() bool {
				return correction != nil
			})
			if !temporal.IsCanceledError(err) && !ok {
				return nil, fmt.Errorf("workflow failed due to bad subscription ID that was not corrected within an hour: %s", request.SubscriptionID)
			}
			cancelCorrection()
			request.SubscriptionID = correction.SubscriptionID

			// recursive call to try again!
			return o.setupJFrog(ctx,
				secrets,
				params,
				request,
				counter+1)
		}
	}
	// unspecified error handling
	return nil, err
}
