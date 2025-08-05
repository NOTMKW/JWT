package main

import (
	"errors"
	"log"

	"github.com/NOTMKW/JWT/internal/config"
	"github.com/NOTMKW/JWT/internal/handler"
	"github.com/NOTMKW/JWT/internal/repo"
	"github.com/NOTMKW/JWT/internal/routes"
	"github.com/NOTMKW/JWT/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				code = fiberErr.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New)
	app.Use(cors.New())

	userRepo := repo.NewUserRepository()

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	authHandler := handler.NewAuthHandler(authService)

	routes.SetupAuthRoutes(app, authHandler)

	log.Printf("server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))

}
