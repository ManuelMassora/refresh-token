package di

import (
	"refresh-token/internal/config"
	"refresh-token/internal/handler"
	"refresh-token/internal/repo"
	"refresh-token/internal/auth"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Container struct {
	UserRepo    *repo.UserRepo
	ItemRepo    *repo.ItemRepo
	SessionRepo *repo.SessionRepo

	JWTMarker *auth.JWTMarker

	AuthHandler *handler.AuthHandler
	ItemHandler *handler.ItemHandler
	UserHandler *handler.UserHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	v := validator.New()

	userRepo := repo.NewUserRepo(db)
	itemRepo := repo.NewItemRepo(db)
	sessionRepo := repo.NewSessionRepo()

	jwtMarker := auth.NewJWTMarker(cfg.JWT_SECRET)

	authHandler := handler.NewAuthHandler(userRepo, sessionRepo, jwtMarker, v)
	itemHandler := handler.NewItemHandler(itemRepo, v)
	userHandler := handler.NewUserHandler(userRepo, v)

	return &Container{
		UserRepo:    userRepo,
		ItemRepo:    itemRepo,
		SessionRepo: sessionRepo,
		JWTMarker:   jwtMarker,
		AuthHandler: authHandler,
		ItemHandler: itemHandler,
		UserHandler: userHandler,
	}
}
