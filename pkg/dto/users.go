package dto

type User struct {
	ID       int
	Login    string
	Name     string
	IsActive bool
}

type CreateUserRequest struct {
	Login    string
	Name     string
	IsActive bool
}

type UpdateUserRequest struct {
	ID       int
	Login    string
	Name     string
	IsActive bool
}
