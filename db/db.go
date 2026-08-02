package db

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"pos-backend/models"
)

var DB *gorm.DB

// Init เชื่อมต่อไฟล์ฐานข้อมูล SQLite และสร้างตารางอัตโนมัติตาม struct ใน models
func Init(path string) {
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = database.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.CategoryOptionTemplate{},
		&models.CategoryOptionTemplateChoice{},
		&models.MenuItem{},
		&models.MenuOptionGroup{},
		&models.MenuOptionChoice{},
		&models.DiningTable{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderItemOption{},
		&models.Payment{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	DB = database
}
