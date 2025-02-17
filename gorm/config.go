package gorm

import (
	"fmt"
	"os"
)

func DSN() string {
	host := mustGetEnv("DB_HOST")
	port := mustGetEnv("DB_PORT")
	sslMode := mustGetEnv("DB_SSL_MODE")
	dbName := mustGetEnv("DB_NAME")
	user := mustGetEnv("DB_USER")
	password := mustGetEnv("DB_PASSWORD")

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbName, port, sslMode)
}

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(fmt.Errorf("%s not set", name))
	}
	return value
}
