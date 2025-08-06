package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"time"

	"github.com/NOTMKW/JWT/internal/dto"
	"github.com/NOTMKW/JWT/internal/model"
	"github.com/NOTMKW/JWT/internal/repo"

	"github.com/golang-jwt/jwt/v5"
)

type Authservice struct {
	userRepo  *repo.UserRepository
	EmailService *EmailService
	jwtSecret []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repo.UserRepository, emailService *EmailService, jwtSecret string) *Authservice {
	return &Authservice{
		userRepo:  userRepo,
		EmailService: emailService,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Authservice) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	role := req.Role
	if role == "" {
		role = model.RoleUser
	}

	if role != model.RoleAdmin && role != model.RoleUser {
		return nil, errors.New("invalid role")
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     role,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *Authservice) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid credentials")
	}

	code, err := s.generateMFACode()
	if err != nil {
		return nil, errors.New("failed to generate MFA code")
	}

	mfaCode := &model.MFACode{
		Email : user.Email,
		Code : code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.StoreMFACode(mfaCode); err != nil {
		return nil, errors.New("failed to store MFA Code")
	}

	if err := s.EmailService.SendMFACode(user.Email, code); err != nil {
		return nil, errors.New("failed to send MFA Code")
	}

	return &dto.AuthResponse{
		RequiredMFA: true,
		Message: "MFA code sent to your email",
	}, nil 
}

func (s *Authservice) VerifyMFACode (req *dto.VerifyMFARequest) (*dto.AuthResponse, error) {
	StoredMFACode, err := s.userRepo.GetMFACode(req.Email)
	if err != nil {
		return nil, errors.New("invalid or expired MFA Code")
	}

	if StoredMFACode.IsExpired() {
		_ = s.userRepo.DeleteMFACode(req.Email)
		return nil, errors.New("MFA Code Expired")
	}

	if StoredMFACode.Code!= req.Code {
		return nil, errors.New("Invalid MFA Code")
	}

	_ = s.userRepo.DeleteMFACode(req.Email)

	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("user not found")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID : user.ID,
			Username: user.Username,
			Email: user.Email,
			Role: user.Role,
		},
	}, nil 
}

func (s *Authservice) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (s *Authservice) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(24 * time.Hour)},
			IssuedAt:  &jwt.NumericDate{time.Now()},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Authservice) generateMFACode() (string, error) {
	code := ""
	for i := 0; 1 < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", num.Int64)
	}
	return code, nil
}
