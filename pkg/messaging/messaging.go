package messaging

type OnboardApp struct {
	UserID         string
	JWT            string
	SubscriptionID string
}

type OnboardJFrogApp struct {
	UserID string
	JWT    string
}
type OnboardSonarQubeApp struct {
}
type GetSecretsRequest struct {
	UserID string
}
type GetSecretsResponse struct {
	Secrets map[string]string
}

type SetupJFrogRequest struct {
	SubscriptionID string
	Secret         string
}

type SetupJFrogResponse struct {
	SubscriptionID string
	Secret         string
}
type RequestCorrectionRequest struct {
	BadSubscriptionID string
	UserID            string
}

type CorrectSubscriptionIDCommand struct {
	SubscriptionID string
}
