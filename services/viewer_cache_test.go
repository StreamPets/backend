package services

import (
	"slices"
	"testing"

	"github.com/streampets/backend/models"
)

func TestGetViewers(t *testing.T) {
	channelName := "channel name"

	cache := NewViewerCacheService()

	got := cache.GetViewers(channelName)
	want := []Pet{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestAddViewer(t *testing.T) {
	channelNameOne := "channel name one"
	channelNameTwo := "channel name two"

	viewerOne := Pet{ViewerId: models.UserId("viewer id one")}
	viewerTwo := Pet{ViewerId: models.UserId("viewer id two")}

	cache := NewViewerCacheService()

	cache.AddViewer(channelNameOne, viewerOne)
	cache.AddViewer(channelNameTwo, viewerTwo)

	viewersOne := cache.GetViewers(channelNameOne)
	viewersTwo := cache.GetViewers(channelNameTwo)

	wantOne := []Pet{viewerOne}
	wantTwo := []Pet{viewerTwo}

	if !slices.Equal(viewersOne, wantOne) {
		t.Errorf("got %s want %s", viewersOne, wantOne)
	}
	if !slices.Equal(viewersTwo, wantTwo) {
		t.Errorf("got %s want %s", viewersTwo, wantTwo)
	}
}

func TestRemoveViewer(t *testing.T) {
	channelName := "channel name"
	viewerId := models.UserId("viewer id")
	viewer := Pet{ViewerId: viewerId}

	cache := NewViewerCacheService()

	cache.AddViewer(channelName, viewer)
	cache.RemoveViewer(channelName, viewer.ViewerId)

	got := cache.GetViewers(channelName)
	want := []Pet{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestUpdateViewer(t *testing.T) {
	channelName := "channel name"
	viewerId := models.UserId("viewer id")
	viewer := Pet{ViewerId: viewerId}
	image := "image"

	cache := NewViewerCacheService()

	cache.AddViewer(channelName, viewer)
	cache.UpdateViewer(channelName, image, viewerId)

	viewers := cache.GetViewers(channelName)

	if len(viewers) != 1 {
		t.Errorf("expected a singleton list but was of length %d", len(viewers))
	}
	if viewers[0].Image != image {
		t.Errorf("got %s want %s", viewers[0].Image, image)
	}
}
