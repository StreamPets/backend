package test

import (
	"github.com/streampets/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(
		&models.ChannelItem{},
		&models.Channel{},
		&models.DefaultChannelItem{},
		&models.Item{},
		&models.OwnedItem{},
		&models.SelectedItem{},
		&models.Viewer{},
	); err != nil {
		panic(err)
	}

	return db
}
