package config

import (
	"encoding/base64"

	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"gorm.io/gorm"
)

func CreateAuthService(db *gorm.DB) *services.AuthService {
	extensionSecret, err := base64.StdEncoding.DecodeString(mustGetEnv("EXTENSION_SECRET"))
	if err != nil {
		panic(err)
	}

	channelRepo := repositories.NewChannelRepo(db)
	authService := services.NewAuthService(channelRepo, string(extensionSecret))

	return authService
}
