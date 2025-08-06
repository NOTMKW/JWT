package handler

import (
	"strings"

	"github.com/NOTMKW/JWT/internal/dto"
	"github.com/NOTMKW/JWT/internal/model"
	"github.com/NOTMKW/JWT/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.Authservice
	validator   *validator.Validate
}

func NewAuthHandler(authService *service.Authservice) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid Response Body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Validation Failed: " + err.Error(),
		})
	}
	response, err := h.authService.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Validation failed: " + err.Error(),
		})
	}

	authResponse, err := h.authService.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(authResponse)

}

func (h *AuthHandler) VerifyMFA(c *fiber.Ctx) error {
	var req dto.VerifyMFARequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid Response Body",
		})
	}

	authResponse, err := h.authService.VerifyMFACode(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error : err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(authResponse)
}

func (h *AuthHandler) Protected(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Authorization header required",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Bearer token required",
		})
	}

	claims, err := h.authService.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Invalid token",
		})
	}

	if claims.Role == model.RoleAdmin {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Admin access granted",
			"user": fiber.Map{
				"id": claims.UserID,
				"email": claims.Email,
				"role": claims.Role,
			},
			"data": "access to all users & admin functions",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User access granted",
		"user": fiber.Map{
			"id":    claims.UserID,
			"email": claims.Email,
			"role": claims.Role,
		},
	})
}
