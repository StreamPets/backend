package config

import (
	"os"

	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"gorm.io/gorm"
)

func CreateAuthService(db *gorm.DB) services.AuthService {
	clientSecret := os.Getenv("CLIENT_SECRET")

	channelRepo := repositories.NewChannelRepo(db)

	return services.NewAuthService(channelRepo, clientSecret)
}
