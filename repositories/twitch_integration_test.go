//go:build integration

package repositories

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func setupTwitchRepository() *TwitchRepository {

	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}

	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	repo, err := NewTwitchRepository(clientId, clientSecret)
	if err != nil {
		panic(err)
	}

	return repo
}

func TestGetUsername(t *testing.T) {
	channelId := models.TwitchId("83125762")
	expected := "ljrexcodes"

	twitchRepo := setupTwitchRepository()

	username, err := twitchRepo.GetUsername(channelId)

	assert.NoError(t, err)
	assert.Equal(t, expected, username)
}

func TestGetUserId(t *testing.T) {
	channelName := "ljrexcodes"
	expected := models.TwitchId("83125762")

	twitchRepo := setupTwitchRepository()

	userId, err := twitchRepo.GetGetId(channelName)

	assert.NoError(t, err)
	assert.Equal(t, expected, userId)
}
