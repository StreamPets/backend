package config

import (
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"gorm.io/gorm"
)

func CreateAuthService(db *gorm.DB) *services.AuthService {
	clientSecret := mustGetEnv("CLIENT_SECRET")

	channelRepo := repositories.NewChannelRepo(db)
	authService := services.NewAuthService(channelRepo, clientSecret)

	return authService
}
