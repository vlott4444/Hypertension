package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"hypertensionstages/internal/app/ds"
	"hypertensionstages/internal/app/dsn"
)

func main() {

	_ = godotenv.Load()

	db, err := gorm.Open(
		postgres.Open(dsn.FromEnv()),
		&gorm.Config{},
	)

	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.HypertensionStage{},
		&ds.User{},
		&ds.HypertensionLike{},
	)

	if err != nil {
		panic("migration failed")
	}
}
