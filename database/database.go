package database

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg config.Database) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Uri), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
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
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}


	return db, nil
}
