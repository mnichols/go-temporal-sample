package messaging

type OnboardApp struct {
	UserID        string
	JWT           string
	ClientIDJfrog string
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
	ClientIDJfrog     string
	ClientSecretJfrog string
	TenantIdJfrog     string
	Jfrogname         string
}

type SetupJFrogResponse struct {
	ClientIDJfrog     string
	ClientSecretJfrog string
	TenantIdJfrog     string
	Jfrogname         string
}
type RequestCorrectionRequest struct {
	BadSubscriptionID string
	UserID            string
	Message           string
}

type CorrectionCommand struct {
	ClientId              *string   `json:"ClientId"`
	ClientSecret          *string   `json:"ClientSecret"`
	TenantId              *string   `json:"TenantId"`
	ClientIdJfrog         *string   `json:"ClientIdJfrog"`
	ClientSecretJfrog     *string   `json:"ClientSecretJfrog"`
	TenantIdJfrog         *string   `json:"TenantIdJfrog"`
	Accountname           *string   `json:"Accountname"`
	Emailid               *string   `json:"Emailid"`
	Jfrogname             *string   `json:"Jfrogname"`
	Packagetype           *string   `json:"Packagetype"`
	Repotype              *string   `json:"Repotype"`
	Groupname             *string   `json:"Groupname"`
	Repositories          *[]string `json:"Repositories"`
	DefaultDeploymentRepo *string   `json:"DefaultDeploymentRepo"`
}
