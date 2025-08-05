package routes

import (
	"github.com/NOTMKW/JWT/internal/handler"

	"github.com/gorilla/mux"
)

func SetupAuthRoutes(router *mux.Router, authHandler *handler.AuthHandler) {
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("api/auth/login", authHandler.Login).Methods("POST")

	router.HandleFunc("/api/auth/protected", authHandler.Protected).Methods("GET")
}