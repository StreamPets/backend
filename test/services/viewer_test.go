package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

func TestGetViewer(t *testing.T) {
	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	channelName := "channel name"
	username := "username"
	image := "image"

	item := models.Item{
		ItemID:  uuid.New(),
		Name:    "name",
		Rarity:  "rarity",
		Image:   image,
		PrevImg: "prev image",
	}

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.Twitcher]()

	mock.When(itemMock.GetSelectedItem(userID, channelID)).ThenReturn(item, nil)
	mock.When(twitchMock.GetUserID(channelName)).ThenReturn(channelID, nil)

	viewerService := services.NewViewerService(itemMock, twitchMock)

	viewer, err := viewerService.GetViewer(userID, channelName, username)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	expected := services.Viewer{
		UserID:   userID,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(userID, channelID)
	if viewer != expected {
		t.Errorf("expected %s got %s", expected, viewer)
	}
}
