package service

import (
	"errors"
	"time"

	"github.com/NOTMKW/JWT/internal/dto"
	"github.com/NOTMKW/JWT/internal/model"
	"github.com/NOTMKW/JWT/internal/repo"

	"github.com/golang-jwt/jwt/v5"
)

type Authservice struct {
	userRepo  *repo.UserRepository
	jwtSecret []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *repo.UserRepository, jwtSecret string) *Authservice {
	return &Authservice{
		userRepo:  userRepo,
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
		Role: 	user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(24 * time.Hour)},
			IssuedAt:  &jwt.NumericDate{time.Now()},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
