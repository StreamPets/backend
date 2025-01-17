package services

import (
	"slices"
	"testing"

	"github.com/streampets/backend/models"
)

func TestGetPets(t *testing.T) {
	channelName := "channel name"

	cache := NewPetCacheService()

	got := cache.GetPets(channelName)
	want := []Pet{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
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

	if !slices.Equal(petsOne, wantOne) {
		t.Errorf("got %s want %s", petsOne, wantOne)
	}
	if !slices.Equal(petsTwo, wantTwo) {
		t.Errorf("got %s want %s", petsTwo, wantTwo)
	}
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

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
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

	if len(pets) != 1 {
		t.Errorf("expected a singleton list but was of length %d", len(pets))
	}
	if pets[0].Image != image {
		t.Errorf("got %s want %s", pets[0].Image, image)
	}
}
