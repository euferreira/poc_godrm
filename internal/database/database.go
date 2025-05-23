package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log"
	"projeto_drm/poc/internal/models"
)

var DB *gorm.DB

func InitDatabase() {
	var err error

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	DB = db

	err = DB.AutoMigrate(&models.Asset{}, &models.ProcessedAsset{})
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	fmt.Println("Database connected and migrated successfully.")
}
