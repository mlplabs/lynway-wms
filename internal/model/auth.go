package model

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrUserInactive       = errors.New("user account is inactive")
)

type TokenClaims struct {
	LocalUserID int    // Заполняется для AuthModeLocal
	ExternalSub string // Заполняется для AuthModeExternal (поле "sub" из внешнего JWT)
	Name        string // Опционально: имя/email для автосоздания локального профиля
}

type UserInfo struct {
	ID           int
	Login        string
	Name         string
	PasswordHash *string
	ExternalSub  *string
	IsActive     bool
}

type LoginRequest struct {
	Login    string `json:"login" example:"picker_04"`
	Password string `json:"password,omitempty" example:"secret123"`
}
