package config

import (
	"github.com/streampets/backend/repositories"
)

func CreateTwitchRepo() (repositories.TwitchRepository, error) {
	clientID := mustGetEnv("CLIENT_ID")
	clientSecret := mustGetEnv("CLIENT_SECRET")

	return repositories.NewTwitchRepository(clientID, clientSecret)
}
