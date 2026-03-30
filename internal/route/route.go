package route

import (
	"refresh-token/internal/di"
	"refresh-token/internal/middlewares"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RegisterRoutes(container *di.Container) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(3 *time.Second))

	r.Post("/login", container.AuthHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.Auth(container.JWTMarker))
		r.Post("/logout", container.AuthHandler.Logout)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", container.UserHandler.CreateUser)
		r.Get("/", container.UserHandler.GetAllUsers)
		r.Get("/{id}", container.UserHandler.GetUser)
		r.Put("/{id}", container.UserHandler.UpdateUser)
		r.Delete("/{id}", container.UserHandler.DeleteUser)
	})

	r.Route("/items", func(r chi.Router) {
		r.Use(middlewares.Auth(container.JWTMarker))
		r.Post("/", container.ItemHandler.CreateItem)
		r.Get("/", container.ItemHandler.GetAllItems)
		r.Get("/{id}", container.ItemHandler.GetItem)
		r.Put("/{id}", container.ItemHandler.UpdateItem)
		r.Delete("/{id}", container.ItemHandler.DeleteItem)
	})

	r.Route("/tokens", func(r chi.Router) {
		r.Route("/renew", func(r chi.Router) {
			r.Post("/", container.AuthHandler.RenewAccessToken)
		})
		r.Route("/revoke/{id}", func(r chi.Router) {
			r.Post("/", container.AuthHandler.RevokeSession)
		})
	})

	return r
}