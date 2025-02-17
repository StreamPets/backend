package gorm

import (
	_ "github.com/lib/pq"
	streampets "github.com/streampets/backend"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	dsn string
}

func NewDB(dsn string) *DB {
	return &DB{
		dsn: dsn,
	}
}

func (db *DB) Open() (err error) {
	db.DB, err = gorm.Open(postgres.Open(db.dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&channelItem{},
		&channel{},
		&defaultChannelItem{},
		&streampets.Item{},
		&ownedItem{},
		&selectedItem{},
	); err != nil {
		return err
	}

	return nil
}
