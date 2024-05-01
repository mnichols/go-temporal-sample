package cicd

import (
	"context"
	"fmt"
	"github.com/mnichols/go-temporal-sample/pkg/messaging"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

var TypeHandlers *Handlers

const RequestParamsErr = "RequestParamsErr"

type Handlers struct {
}

func (h *Handlers) SetupJFrog(ctx context.Context, req *messaging.SetupJFrogRequest) (*messaging.SetupJFrogResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Debug("setting up jfrog")
	if req.ClientIDJfrog == "BAD_JFROG_CLIENT_ID" {
		err := fmt.Errorf("Jfrog client id %s is bad", req.ClientIDJfrog)
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), RequestParamsErr, err)
	}

	return &messaging.SetupJFrogResponse{
		ClientIDJfrog:     req.ClientIDJfrog,
		ClientSecretJfrog: req.ClientSecretJfrog,
		TenantIdJfrog:     req.TenantIdJfrog,
		Jfrogname:         req.Jfrogname,
	}, nil
}
