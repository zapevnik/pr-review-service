package httpserver

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func UseMiddlewares(r chi.Router) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}
