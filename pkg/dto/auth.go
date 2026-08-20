package dto

type LoginUserRequest struct {
	Email    string `json:"email"    example:"user@localhost.local"` // Email пользователя
	Password string `json:"password" example:"password"`             // Пароль пользователя
}

type LoginUserResponse struct {
	AccessToken  string `json:"access_token"`  // Access токен
	RefreshToken string `json:"refresh_token"` // refresh токен
}

type RefreshUserRequest struct {
	RefreshToken string `json:"refresh_token"` // refresh токен
}
type RefreshUserResponse struct {
	AccessToken  string `json:"access_token"`  // Access токен
	RefreshToken string `json:"refresh_token"` // refresh токен
}
