package main

import (
	"context"
	"log"
	"net/http"
	"refresh-token/internal/config"
	"refresh-token/internal/db"
	"refresh-token/internal/di"
	"refresh-token/internal/infra/redis"
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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Erro ao obter instância SQL: %v", err)
	}
	defer sqlDB.Close()

	redis.InitRedis()
	if err := redis.PingRedis(context.Background()); err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}
	log.Println("Conexão com Redis estabelecida com sucesso!")

	container := di.NewContainer(db, cfg)

	r := route.RegisterRoutes(container)
	log.Printf("Servidor rodando na porta %s", cfg.SERVER_PORT)
	if err := http.ListenAndServe(":"+cfg.SERVER_PORT, r); err != nil { //http.ListenAndServeTLS em Producao
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
