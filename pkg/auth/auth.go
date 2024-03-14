package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mnichols/go-temporal-sample/pkg/messaging"
	"io"
	"net/http"
)

var TypeHandlers *Handlers

type Handlers struct {
}

func (h *Handlers) GetVaultSecrets(ctx context.Context, cmd *messaging.GetSecretsRequest) (*messaging.GetSecretsResponse, error) {

	res, err := http.Get("https://httpbin.org/get?foo=bar")
	if err != nil {
		return nil, fmt.Errorf("failed to get data %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not call api")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	output := map[string]interface{}{}
	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, err
	}
	secrets := map[string]string{}
	secrets["jfrog"] = "123"
	result, exists := output["args"]
	if exists {
		if args, ok := result.(map[string]interface{}); ok {
			secrets["foo"] = args["foo"].(string)
		}
	}

	return &messaging.GetSecretsResponse{Secrets: secrets}, nil
}
