package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"ssr_immemory/internal/app/ds"
	"ssr_immemory/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&ds.Paint{},
		&ds.RequestsPaint{},
		&ds.PaintRequest{},
		&ds.User{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}