package db

import (
	"fmt"
	"log"
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

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("erro no AutoMigrate: %w", err)
	}

	if err := insertInitialRoles(db); err != nil {
		return nil, fmt.Errorf("erro ao inserir roles iniciais: %w", err)
	}

	log.Println("Conexão com o banco de dados e migrações realizadas com sucesso!")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.Item{},
		&model.Device{},
	)
}

func insertInitialRoles(db *gorm.DB) error {
	const insertRole = `
		INSERT INTO roles (id, name) VALUES
			(1, 'ADMIN'),
			(2, 'BASIC')
		ON CONFLICT (id) DO NOTHING;
	`
	if err := db.Exec(insertRole).Error; err != nil {
		return err
	}
	return nil
}