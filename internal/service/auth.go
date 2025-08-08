
package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/NOTMKW/JWT/internal/config"
	"github.com/NOTMKW/JWT/internal/dto"
	"github.com/NOTMKW/JWT/internal/model"
	"github.com/NOTMKW/JWT/internal/repo"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo     *repo.UserRepository
	emailService *EmailService
	config       *config.Config
	jwtSecret    []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewAuthService(userRepo *repo.UserRepository, emailService *EmailService, cfg *config.Config, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		emailService: emailService,
		config:       cfg,
		jwtSecret:    []byte(jwtSecret),
	}
}

func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
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

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
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
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.StoreMFACode(mfaCode); err != nil {
		return nil, errors.New("failed to store MFA code")
	}

	if err := s.emailService.SendMFACode(user.Email, code); err != nil {
		return nil, errors.New("failed to send MFA code")
	}

	return &dto.AuthResponse{
		RequiredMFA: true,
		Message:     "MFA code sent to your email",
	}, nil
}

func (s *AuthService) VerifyMFACode(req *dto.VerifyMFARequest) (*dto.AuthResponse, error) {
	storedMFACode, err := s.userRepo.GetMFACode(req.Email)
	if err != nil {
		return nil, errors.New("invalid or expired MFA code")
	}

	if storedMFACode.IsExpired() {
		_ = s.userRepo.DeleteMFACode(req.Email)
		return nil, errors.New("MFA code expired")
	}

	if storedMFACode.Code != req.Code {
		return nil, errors.New("invalid MFA code")
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
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) GoogleAuth(code string) (*dto.AuthResponse, error) {
	token, err := s.exchangeCodeForToken(code)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.CreateOrUpdateGoogleUser(userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		return nil, err
	}

	jwtToken, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: jwtToken,
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
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

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(24 * time.Hour)},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateMFACode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", num.Int64())
	}
	return code, nil
}

func (s *AuthService) exchangeCodeForToken(code string) (*GoogleTokenResponse, error) {
	data := url.Values{
		"client_id":     {s.config.GoogleClientID},
		"client_secret": {s.config.GoogleClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {s.config.GoogleRedirectURL},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *AuthService) getUserInfoFromGoogle(accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
