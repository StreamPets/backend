package config

import (
	"encoding/base64"

	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

func CreateAuthService(channelRepo *repositories.ChannelRepo) *services.AuthService {
	extensionSecret, err := base64.StdEncoding.DecodeString(mustGetEnv("EXTENSION_SECRET"))
	if err != nil {
		panic(err)
	}

	return services.NewAuthService(channelRepo, string(extensionSecret))
}
