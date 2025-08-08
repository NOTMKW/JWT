package routes

import (
	"github.com/NOTMKW/JWT/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App, authHandler *handler.AuthHandler) {
	api := app.Group("/api")
	auth := api.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/verify-mfa", authHandler.VerifyMFA)
	auth.Post("/google", authHandler.GoogleAuth)
	
	auth.Get("/protected", authHandler.Protected)
}