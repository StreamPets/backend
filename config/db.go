package config

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/streampets/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	host := mustGetEnv("DB_HOST")
	port := mustGetEnv("DB_PORT")
	sslMode := mustGetEnv("DB_SSL_MODE")
	dbName := mustGetEnv("DB_NAME")
	user := mustGetEnv("DB_USER")
	password := mustGetEnv("DB_PASSWORD")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbName, port, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
