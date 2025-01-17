package config

import (
	"github.com/streampets/backend/repositories"
)

func CreateTwitchRepo() *repositories.TwitchRepository {
	clientId := mustGetEnv("CLIENT_ID")
	clientSecret := mustGetEnv("CLIENT_SECRET")

	twitchRepo, err := repositories.NewTwitchRepository(clientId, clientSecret)
	if err != nil {
		panic(err)
	}

	return twitchRepo
}
