package dto

type SystemInfoResponse struct {
	Version         string `json:"version"`
	AuthMode        string `json:"auth_mode"`
	RequirePassword bool   `json:"require_password"`
}
