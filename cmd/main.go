package main

import (
	"errors"
	"log"
	"strings"

	"github.com/NOTMKW/JWT/internal/config"
	"github.com/NOTMKW/JWT/internal/handler"
	"github.com/NOTMKW/JWT/internal/model"
	"github.com/NOTMKW/JWT/internal/repo"
	"github.com/NOTMKW/JWT/internal/routes"
	"github.com/NOTMKW/JWT/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	db, err := initDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.MFACode{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

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

	app.Use(logger.New())
	app.Use(cors.New())

	userRepo := repo.NewUserRepository(db)

	EmailService := *service.NewEmailService(cfg)

	AuthService := service.NewAuthService(userRepo, &EmailService, cfg, cfg.JWTSecret)

	authHandler := handler.NewAuthHandler(AuthService)

	routes.SetupAuthRoutes(app, authHandler)

	log.Printf("server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}

func initDatabase(databaseURL string) (*gorm.DB, error) {
	if strings.HasPrefix(databaseURL, "postgresql://") {
		return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	} else if strings.HasPrefix(databaseURL, "sqlite://") {
		dbPath := strings.TrimPrefix(databaseURL, "sqlite://")
		return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}
	return nil, errors.New("Unsupported database type")
}
