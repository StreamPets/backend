package services

import (
	"testing"

	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestGetPets(t *testing.T) {
	channelId := models.TwitchId("channel id")

	cache := NewPetCacheService()

	got := cache.GetPets(channelId)
	want := []Pet{}

	assert.Equal(t, want, got)
}

func TestAddPet(t *testing.T) {
	channelIdOne := models.TwitchId("channel id one")
	channelIdTwo := models.TwitchId("channel id two")

	petOne := Pet{UserId: models.TwitchId("user id one")}
	petTwo := Pet{UserId: models.TwitchId("user id two")}

	cache := NewPetCacheService()

	cache.AddPet(channelIdOne, petOne)
	cache.AddPet(channelIdTwo, petTwo)

	petsOne := cache.GetPets(channelIdOne)
	petsTwo := cache.GetPets(channelIdTwo)

	wantOne := []Pet{petOne}
	wantTwo := []Pet{petTwo}

	assert.Equal(t, wantOne, petsOne)
	assert.Equal(t, wantTwo, petsTwo)
}

func TestRemovePet(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	pet := Pet{UserId: userId}

	cache := NewPetCacheService()

	cache.AddPet(channelId, pet)
	cache.RemovePet(channelId, pet.UserId)

	got := cache.GetPets(channelId)
	want := []Pet{}

	assert.Equal(t, want, got)
}

func TestUpdatePet(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	pet := Pet{UserId: userId}
	image := "image"

	cache := NewPetCacheService()

	cache.AddPet(channelId, pet)
	cache.UpdatePet(channelId, userId, image)

	pets := cache.GetPets(channelId)

	if assert.Equal(t, 1, len(pets)) {
		assert.Equal(t, image, pets[0].Image)
	}
}
