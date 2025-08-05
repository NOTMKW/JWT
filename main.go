package main

import (
	"log"
	"net/http"

	"github.com/NOTMKW/JWT/internal/config"
	"github.com/NOTMKW/JWT/internal/handler"
	"github.com/NOTMKW/JWT/internal/repo"
	"github.com/NOTMKW/JWT/internal/routes"
	"github.com/NOTMKW/JWT/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	userRepo := repo.NewUserRepository()

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	authHandler := handler.NewAuthHandler(authService)

	router := mux.NewRouter()

	routes.SetupAuthRoutes(router, authHandler)

	log.Printf("server starting on porter %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
