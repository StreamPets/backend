package gorm

import (
	streampets "github.com/streampets/backend"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateTestDB() (*DB, error) {
	db := &DB{}
	var err error

	db.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&channelItem{},
		&channel{},
		&defaultChannelItem{},
		&streampets.Item{},
		&ownedItem{},
		&selectedItem{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
