package services

import (
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestGetViewer(t *testing.T) {
	mock.SetUp(t)

	viewerId := models.TwitchId("viewer id")
	channelId := models.TwitchId("channel id")
	username := "username"
	image := "image"
	item := models.Item{Image: image}

	itemMock := mock.Mock[ItemRepo]()
	mock.When(itemMock.GetSelectedItem(viewerId, channelId)).ThenReturn(item, nil)

	viewerService := NewPetService(itemMock)

	viewer, err := viewerService.GetPet(viewerId, channelId, username)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	expected := Pet{
		ViewerId: viewerId,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(viewerId, channelId)
	if viewer != expected {
		t.Errorf("expected %s got %s", expected, viewer)
	}
}
