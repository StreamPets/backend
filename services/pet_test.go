package services

import (
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	mock.SetUp(t)

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	username := "username"
	image := "image"
	item := models.Item{Image: image}

	itemMock := mock.Mock[SelectedItemGetter]()
	mock.When(itemMock.GetSelectedItem(userId, channelId)).ThenReturn(item, nil)

	petService := NewPetService(itemMock)

	pet, err := petService.GetPet(userId, channelId, username)

	expected := Pet{
		UserId:   userId,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(userId, channelId)

	assert.NoError(t, err)
	assert.Equal(t, expected, pet)
}
