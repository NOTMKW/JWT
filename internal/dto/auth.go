package dto

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email" validate:"required,email"`
}

type LoginRequest struct {
	Email	string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User UserResponse `json:"user"`
}

type UserResponse struct {
	ID       string   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}