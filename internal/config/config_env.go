package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DSN    string
	JWT_SECRET   string
	SERVER_PORT string
	REDIS_ADDR string
}

func LoadConfig(envPath string) (*Config, error) {

	err := godotenv.Load(envPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("erro ao carregar arquivo .env: %w", err)
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		log.Printf("SERVER_PORT not set, using default: %s", serverPort)
	}

	dbDSN := os.Getenv("DSN")
	if dbDSN == "" {
		log.Printf("DATABASE_DSN not set, using default: %s", dbDSN)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("JWT_SECRET not set, using default: %s", jwtSecret)
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Printf("REDIS_ADDR not set, using default: %s", redisAddr)
	}

	cfg := &Config{
		DSN:    dbDSN,
		JWT_SECRET:   jwtSecret,
		SERVER_PORT: serverPort,
		REDIS_ADDR: redisAddr,
	}
	return cfg, nil
}