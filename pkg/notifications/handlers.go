package notifications

import (
	"context"
	"github.com/mnichols/go-temporal-sample/pkg/messaging"
	"go.temporal.io/sdk/activity"
)

var TypeHandlers *Handlers

type Handlers struct {
}

func (h *Handlers) RequestCorrection(ctx context.Context, request messaging.RequestCorrectionRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Debug("sending an email to request a correction", "subscriptionID", request.BadSubscriptionID)
	return nil
}
