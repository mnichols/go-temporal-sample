package cicd

import (
	"context"
	"fmt"
	"github.com/mnichols/go-temporal-sample/pkg/messaging"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

var TypeHandlers *Handlers

const BadSubscriptionIDErr = "Bad Subscription ID"

type Handlers struct {
}

func (h *Handlers) SetupJFrog(ctx context.Context, req *messaging.SetupJFrogRequest) (*messaging.SetupJFrogResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Debug("setting up jfrog")
	if req.SubscriptionID == "GarbageSubscriptionID" {
		err := fmt.Errorf("subscriptionID %s is garbage", req.SubscriptionID)
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), BadSubscriptionIDErr, err)
	}

	return &messaging.SetupJFrogResponse{
		SubscriptionID: req.SubscriptionID,
		Secret:         req.Secret,
	}, nil
}
