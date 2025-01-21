package services

import (
	"testing"

	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestGetPets(t *testing.T) {
	channelName := "channel name"

	cache := NewPetCacheService()

	got := cache.GetPets(channelName)
	want := []Pet{}

	assert.Equal(t, want, got)
}

func TestAddPet(t *testing.T) {
	channelNameOne := "channel name one"
	channelNameTwo := "channel name two"

	petOne := Pet{UserId: models.TwitchId("user id one")}
	petTwo := Pet{UserId: models.TwitchId("user id two")}

	cache := NewPetCacheService()

	cache.AddPet(channelNameOne, petOne)
	cache.AddPet(channelNameTwo, petTwo)

	petsOne := cache.GetPets(channelNameOne)
	petsTwo := cache.GetPets(channelNameTwo)

	wantOne := []Pet{petOne}
	wantTwo := []Pet{petTwo}

	assert.Equal(t, wantOne, petsOne)
	assert.Equal(t, wantTwo, petsTwo)
}

func TestRemovePet(t *testing.T) {
	channelName := "channel name"
	userId := models.TwitchId("user id")
	pet := Pet{UserId: userId}

	cache := NewPetCacheService()

	cache.AddPet(channelName, pet)
	cache.RemovePet(channelName, pet.UserId)

	got := cache.GetPets(channelName)
	want := []Pet{}

	assert.Equal(t, want, got)
}

func TestUpdatePet(t *testing.T) {
	channelName := "channel name"
	userId := models.TwitchId("user id")
	pet := Pet{UserId: userId}
	image := "image"

	cache := NewPetCacheService()

	cache.AddPet(channelName, pet)
	cache.UpdatePet(channelName, image, userId)

	pets := cache.GetPets(channelName)

	if assert.Equal(t, 1, len(pets)) {
		assert.Equal(t, image, pets[0].Image)
	}
}
