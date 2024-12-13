package config

import (
	"github.com/streampets/backend/repositories"
)

func CreateTwitchRepo() *repositories.TwitchRepository {
	clientID := mustGetEnv("CLIENT_ID")
	clientSecret := mustGetEnv("CLIENT_SECRET")

	twitchRepo, err := repositories.NewTwitchRepository(clientID, clientSecret)
	if err != nil {
		panic(err)
	}

	return twitchRepo
}
