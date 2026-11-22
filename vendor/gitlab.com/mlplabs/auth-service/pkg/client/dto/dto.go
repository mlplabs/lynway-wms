package dto

const (
	ClaimUserID  = "user_id"
	ClaimOwnerID = "owner_id"
	ClaimRights  = "rights"
)

type LoginRequest struct {
	Email    string `json:"email"`             // Логин
	Password string `json:"password"`          // Пароль
	ApiKey   string `json:"api_key,omitempty"` // ApiKey, если авторизация не по паролю
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type MeResponse struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	OwnerId int    `json:"owner_id"`
}

type OwnerRegisterRequest struct {
	Name            string `json:"name"`
	Key             string `json:"key"`
	TokenLifetime   int    `json:"token_lifetime"`
	RefreshLifetime int    `json:"refresh_lifetime"`
}

type OwnerRegisterResponse struct {
}
