package db

import (
	"fmt"
	// "log"
	"refresh-token/internal/model"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter sql.DB do GORM: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// if err := autoMigrate(db); err != nil {
	// 	return nil, fmt.Errorf("erro no AutoMigrate: %w", err)
	// }

	// log.Println("Conexão com o banco de dados e migrações realizadas com sucesso!")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Item{},
		&model.Session{},
	)
}
