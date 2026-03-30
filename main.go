package main

import (
	"log"
	"net/http"
	"refresh-token/internal/config"
	"refresh-token/internal/db"
	"refresh-token/internal/di"
	"refresh-token/internal/route"
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

	container := di.NewContainer(db, cfg)

	r := route.RegisterRoutes(container)
	log.Printf("Servidor rodando na porta %s", cfg.SERVER_PORT)
	if err := http.ListenAndServe(":"+cfg.SERVER_PORT, r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
