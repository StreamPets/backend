package services

import (
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestGetViewer(t *testing.T) {
	mock.SetUp(t)

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	username := "username"
	image := "image"
	item := models.Item{Image: image}

	itemMock := mock.Mock[ItemRepo]()
	mock.When(itemMock.GetSelectedItem(userID, channelID)).ThenReturn(item, nil)

	viewerService := NewViewerService(itemMock)

	viewer, err := viewerService.GetViewer(userID, channelID, username)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	expected := Viewer{
		UserID:   userID,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(userID, channelID)
	if viewer != expected {
		t.Errorf("expected %s got %s", expected, viewer)
	}
}
