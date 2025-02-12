package config

import (
	"encoding/base64"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	HttpPort uint16 `envconfig:"HTTP_PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
	ExtensionSecret string `envconfig:"EXTENSION_SECRET" required:"true"`
	Database Database
}

type Database struct {
	Uri string `envconfig:"DB_URI" default:"postgres://postgres:postgres@localhost:5432?sslmode=disable"`
}


func Get() (*AppConfig, error) {
	cfg := new(AppConfig)
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	plainSecret, err := base64.StdEncoding.DecodeString(cfg.ExtensionSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode extension secret: %w", err)
	}

	cfg.ExtensionSecret = string(plainSecret)

	return cfg, nil
}

