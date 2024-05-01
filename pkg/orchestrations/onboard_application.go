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

type OnboardingState struct {
	Correction        *messaging.CorrectionCommand
	RequestCorrection *messaging.RequestCorrectionRequest
}

func (o *Orchestrations) OnboardApplication(ctx workflow.Context, params *messaging.OnboardApp) error {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 3,
	})

	onboardingState := &OnboardingState{}

	logger := workflow.GetLogger(ctx)

	// general query handler for convenience to see the status of the onboarding
	if err := workflow.SetQueryHandler(ctx, QueryOnboardingState, func() (*OnboardingState, error) {
		return onboardingState, nil
	}); err != nil {
		return fmt.Errorf("error setting up query %w", err)
	}

	// correction notification handler
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			workflow.Await(ctx, func() bool {
				return onboardingState.RequestCorrection != nil
			})
			// fire and forget notification
			if err := workflow.ExecuteActivity(ctx, notifications.TypeHandlers.RequestCorrection, onboardingState.RequestCorrection).Get(ctx, nil); err != nil {
				logger.Error("failed to request correction", "err", err)
				// what do we do when we cannot request correction??
			}
			// clear the request
			onboardingState.RequestCorrection = nil
		}
	})

	// correction response/signal handler
	chanCtx, cancelCorrection := workflow.WithCancel(ctx)
	correctionSignalChan := workflow.GetSignalChannel(chanCtx, SignalCorrection)
	workflow.Go(chanCtx, func(ctx workflow.Context) {
		for {
			// pick correction signals off and wait until they are handled and cleared before collecting the next
			correctionSignalChan.Receive(ctx, &onboardingState.Correction)
			logger.Info("received correction", "data", onboardingState.Correction)
			workflow.Await(ctx, func() bool {
				return onboardingState.Correction == nil
			})
		}
	})

	var getSecretsResponse *messaging.GetSecretsResponse

	var jfrogRequest *messaging.SetupJFrogRequest
	var jfrogResponse *messaging.SetupJFrogResponse

	if err := workflow.ExecuteActivity(ctx, auth.TypeHandlers.GetVaultSecrets, &messaging.GetSecretsRequest{
		UserID: params.UserID,
	}).Get(ctx, &getSecretsResponse); err != nil {
		return fmt.Errorf("failed to get secrets %w", err)
	}

	jfrogRequest = &messaging.SetupJFrogRequest{
		ClientIDJfrog:     params.ClientIDJfrog,
		ClientSecretJfrog: getSecretsResponse.Secrets["jfrog"],
	}
	var err error
	if jfrogResponse, err = o.setupJFrog(
		ctx,
		onboardingState,
		getSecretsResponse,
		params,
		jfrogRequest,
		1,
	); err != nil {
		return fmt.Errorf("could not setup jfrog %w", err)
	}

	// no corrections should be received anymore
	// you might want to drain off the signal handler and make sure though...up to you
	cancelCorrection()
	logger.Debug("jfrog setup %v", jfrogResponse)
	return nil
}

func (o *Orchestrations) setupJFrog(ctx workflow.Context,
	state *OnboardingState,
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
		logger.Info("jfrog has been setup")
		return response, nil
	}

	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		switch appErr.Type() {
		case cicd.RequestParamsErr:
			state.RequestCorrection = &messaging.RequestCorrectionRequest{
				Message: err.Error(),
				UserID:  params.UserID,
			}
			ok, cerr := workflow.AwaitWithTimeout(ctx, time.Minute*30, func() bool {
				return state.Correction != nil
			})
			if !ok || cerr != nil {
				return nil, fmt.Errorf("request for correction of Jfrog never arrived")
			}
			request = tryMergeJfrogCorrection(request, state.Correction)
			logger.Info("merged correction with previous request, then retrying", "data", request)
			// clean up the correction we received...
			state.Correction = nil
			return o.setupJFrog(ctx, state, secrets, params, request, counter+1)

			// this timeout can be as long as your secrets would be valid
		}
	}
	// unspecified error handling
	return nil, err
}

func tryMergeJfrogCorrection(request *messaging.SetupJFrogRequest, correction *messaging.CorrectionCommand) *messaging.SetupJFrogRequest {
	if correction.Jfrogname != nil {
		request.Jfrogname = *correction.Jfrogname
	}
	if correction.TenantIdJfrog != nil {
		request.TenantIdJfrog = *correction.TenantIdJfrog
	}
	if correction.ClientSecretJfrog != nil {
		request.ClientSecretJfrog = *correction.ClientSecretJfrog
	}
	if correction.ClientIdJfrog != nil {
		request.ClientIDJfrog = *correction.ClientIdJfrog
	}
	return request
}
