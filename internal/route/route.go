package route

import (
	"refresh-token/internal/handler"
	"refresh-token/internal/middlewares"
	"refresh-token/internal/token"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RegisterRoutes(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	itemHandler *handler.ItemHandler,
	jwt *token.JWTMarker) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/", userHandler.GetAllUsers)
		r.Get("/{id}", userHandler.GetUser)
		r.Put("/{id}", userHandler.UpdateUser)
		r.Delete("/{id}", userHandler.DeleteUser)
	})

	r.Route("/items", func(r chi.Router) {
		r.Use(middlewares.Auth(jwt))
		r.Post("/", itemHandler.CreateItem)
		r.Get("/", itemHandler.GetAllItems)
		r.Get("/{id}", itemHandler.GetItem)
		r.Put("/{id}", itemHandler.UpdateItem)
		r.Delete("/{id}", itemHandler.DeleteItem)
	})

	r.Route("/tokens", func(r chi.Router) {
		r.Route("/renew", func(r chi.Router) {
			r.Post("/", authHandler.RenewAccessToken)
		})
		r.Route("/revoke/{id}", func(r chi.Router) {
			r.Post("/", authHandler.RevokeSession)
		})
	})

	return r
}