package dto

type LoginRequest struct {
	Login    string `json:"login" example:"picker_04"`
	Password string `json:"password,omitempty" example:"secret123"`
}

type LoginUserDTO struct {
	ID   int    `json:"id" example:"4"`
	Name string `json:"name" example:"Иван Сборщик"`
}

type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  LoginUserDTO `json:"user"`
}

////////////////////////////////////////

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
