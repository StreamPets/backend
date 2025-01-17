package services

import (
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestGetUser(t *testing.T) {
	mock.SetUp(t)

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	username := "username"
	image := "image"
	item := models.Item{Image: image}

	itemMock := mock.Mock[ItemRepo]()
	mock.When(itemMock.GetSelectedItem(userId, channelId)).ThenReturn(item, nil)

	petService := NewPetService(itemMock)

	pet, err := petService.GetPet(userId, channelId, username)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	expected := Pet{
		UserId:   userId,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(userId, channelId)
	if pet != expected {
		t.Errorf("expected %s got %s", expected, pet)
	}
}
