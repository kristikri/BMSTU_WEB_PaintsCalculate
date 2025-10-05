package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"ssr_immemory/internal/app/ds"
)

func main() {
	_ = godotenv.Load()
	
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=paint_calculator port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Paint{},
		&ds.PaintRequest{},
		&ds.RequestPaint{},
	)
	if err != nil {
		log.Fatal("cant migrate db:", err)
	}

	log.Println("Migration completed successfully!")
}