package main

import (
	"log"
	"net/http"
	"refresh-token/internal/config"
	"refresh-token/internal/db"
	"refresh-token/internal/handler"
	"refresh-token/internal/repo"
	"refresh-token/internal/route"
	"refresh-token/internal/token"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	db, err := db.InitDB(cfg.DSN)
	if err != nil {
		log.Fatalf("Erro ao inicializar o banco de dados: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	userRepo := repo.NewUserRepo(db)
	itemRepo := repo.NewItemRepo(db)
	sessionRepo := repo.NewSessionRepo(db)

	tokenMarker := token.NewJWTMarker(cfg.JWT_SECRET)

	authHandler := handler.NewAuthHandler(userRepo, sessionRepo, tokenMarker)
	userHandler := handler.NewUserHandler(userRepo)
	itemHandler := handler.NewItemHandler(itemRepo)
	r := route.RegisterRoutes(authHandler, userHandler, itemHandler)
	log.Printf("Servidor rodando na porta %s", cfg.SERVER_PORT)
	if err := http.ListenAndServe(":"+cfg.SERVER_PORT, r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}