package dto

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type VerifyMFARequest struct {
	Email string `json:"email" validate:"required.email`
	Code  string `json:"code"  validate:"required,len=6`
}

type GoogleAuthRequest struct {
	Code string `json:"code" validate "required"`
}

type AuthResponse struct {
	Token       string       `json:"token"`
	User        UserResponse `json:"user"`
	RequiredMFA bool         `json:"requires_mfa.omitempty"`
	Message     string       `json:"message,omitempty"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
