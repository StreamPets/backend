package gorm

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	streampets "github.com/streampets/backend"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	dsn string
}

// TODO: Not sure where to put this, maybe in config?
func GET_DSN() string {
	host := mustGetEnv("DB_HOST")
	port := mustGetEnv("DB_PORT")
	sslMode := mustGetEnv("DB_SSL_MODE")
	dbName := mustGetEnv("DB_NAME")
	user := mustGetEnv("DB_USER")
	password := mustGetEnv("DB_PASSWORD")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbName, port, sslMode)
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

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(fmt.Errorf("%s not set", name))
	}
	return value
}
