package route

import (
	"refresh-token/internal/di"
	"refresh-token/internal/middlewares"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func RegisterRoutes(container *di.Container) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(3 *time.Second))
	r.Use(httprate.LimitByIP(10, time.Minute))

	r.Route("/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RateLimitMiddleware)
			r.Post("/login", container.AuthHandler.Login)
			r.Post("/renew", container.AuthHandler.RenewAccessToken)
			r.Post("/revoke/{id}", container.AuthHandler.RevokeSession)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.Auth(container.JWTMarker))
		r.Post("/logout", container.AuthHandler.Logout)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", container.UserHandler.CreateUser)
		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth(container.JWTMarker))
			r.Use(middlewares.HasAnyRole(container.AuthHandler, "ADMIN"))
			r.Get("/", container.UserHandler.GetAllUsers)
			r.Get("/{id}", container.UserHandler.GetUser)
			r.Put("/{id}/name", container.UserHandler.UpdateUserName)
			r.Put("/{id}/password", container.UserHandler.UpdateUserPassword)
			r.Put("/{id}/role", container.UserHandler.UpdateUserRole)
			r.Delete("/{id}", container.UserHandler.DeleteUser)
		})
	})

	r.Route("/items", func(r chi.Router) {
		r.Use(middlewares.Auth(container.JWTMarker))
		r.Group(func(r chi.Router) {
			r.Use(middlewares.HasAnyRole(container.AuthHandler, "ADMIN"))
			r.Post("/", container.ItemHandler.CreateItem)
			r.Put("/{id}", container.ItemHandler.UpdateItem)
			r.Delete("/{id}", container.ItemHandler.DeleteItem)
		})
		r.Get("/", container.ItemHandler.GetAllItems)
		r.Get("/{id}", container.ItemHandler.GetItem)
	})

	return r
}