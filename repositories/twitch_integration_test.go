//go:build integration

package repositories

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/streampets/backend/models"
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
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if username != expected {
		t.Errorf("expected %s got %s", expected, username)
	}
}

func TestGetViewerId(t *testing.T) {
	channelName := "ljrexcodes"
	expected := models.TwitchId("83125762")

	twitchRepo := setupTwitchRepository()

	viewerId, err := twitchRepo.GetViewerId(channelName)
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if viewerId != expected {
		t.Errorf("expected %s got %s", expected, viewerId)
	}
}
