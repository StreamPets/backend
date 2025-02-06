package test

import (
	"net/http/httptest"

	"github.com/streampets/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CloseNotifierResponseWriter struct {
	*httptest.ResponseRecorder
}

func (c *CloseNotifierResponseWriter) CloseNotify() <-chan bool {
	return make(<-chan bool)
}

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
		&models.User{},
	); err != nil {
		panic(err)
	}

	return db
}
