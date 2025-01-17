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

	viewerOne := Pet{ViewerId: models.UserId("viewer id one")}
	viewerTwo := Pet{ViewerId: models.UserId("viewer id two")}

	cache := NewPetCacheService()

	cache.AddPet(channelNameOne, viewerOne)
	cache.AddPet(channelNameTwo, viewerTwo)

	viewersOne := cache.GetPets(channelNameOne)
	viewersTwo := cache.GetPets(channelNameTwo)

	wantOne := []Pet{viewerOne}
	wantTwo := []Pet{viewerTwo}

	if !slices.Equal(viewersOne, wantOne) {
		t.Errorf("got %s want %s", viewersOne, wantOne)
	}
	if !slices.Equal(viewersTwo, wantTwo) {
		t.Errorf("got %s want %s", viewersTwo, wantTwo)
	}
}

func TestRemovePet(t *testing.T) {
	channelName := "channel name"
	viewerId := models.UserId("viewer id")
	viewer := Pet{ViewerId: viewerId}

	cache := NewPetCacheService()

	cache.AddPet(channelName, viewer)
	cache.RemovePet(channelName, viewer.ViewerId)

	got := cache.GetPets(channelName)
	want := []Pet{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestUpdatePet(t *testing.T) {
	channelName := "channel name"
	viewerId := models.UserId("viewer id")
	viewer := Pet{ViewerId: viewerId}
	image := "image"

	cache := NewPetCacheService()

	cache.AddPet(channelName, viewer)
	cache.UpdatePet(channelName, image, viewerId)

	viewers := cache.GetPets(channelName)

	if len(viewers) != 1 {
		t.Errorf("expected a singleton list but was of length %d", len(viewers))
	}
	if viewers[0].Image != image {
		t.Errorf("got %s want %s", viewers[0].Image, image)
	}
}
