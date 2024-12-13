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

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	repo, err := NewTwitchRepository(clientID, clientSecret)
	if err != nil {
		panic(err)
	}

	return repo
}

func TestGetUsername(t *testing.T) {
	channelID := models.TwitchID("83125762")
	expected := "ljrexcodes"

	twitchRepo := setupTwitchRepository()

	username, err := twitchRepo.GetUsername(channelID)
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if username != expected {
		t.Errorf("expected %s got %s", expected, username)
	}
}

func TestGetUserID(t *testing.T) {
	channelName := "ljrexcodes"
	expected := models.TwitchID("83125762")

	twitchRepo := setupTwitchRepository()

	userID, err := twitchRepo.GetUserID(channelName)
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if userID != expected {
		t.Errorf("expected %s got %s", expected, userID)
	}
}
