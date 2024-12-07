package repositories_test

import (
	"github.com/streampets/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createTestDB() *gorm.DB {
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
		&models.Schedule{},
		&models.SelectedItem{},
		&models.User{},
	); err != nil {
		panic(err)
	}

	return db
}
