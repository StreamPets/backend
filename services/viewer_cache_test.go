package services

import (
	"slices"
	"testing"

	"github.com/streampets/backend/models"
)

func TestGetViewers(t *testing.T) {
	channelID := models.TwitchID("channel id")

	cache := NewViewerCacheService()

	got := cache.GetViewers(channelID)
	want := []Viewer{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestAddViewer(t *testing.T) {
	channelIdOne := models.TwitchID("channel id one")
	channelIdTwo := models.TwitchID("channel id two")

	viewerOne := Viewer{UserID: models.TwitchID("user id one")}
	viewerTwo := Viewer{UserID: models.TwitchID("user id two")}

	cache := NewViewerCacheService()

	cache.AddViewer(channelIdOne, viewerOne)
	cache.AddViewer(channelIdTwo, viewerTwo)

	viewersOne := cache.GetViewers(channelIdOne)
	viewersTwo := cache.GetViewers(channelIdTwo)

	wantOne := []Viewer{viewerOne}
	wantTwo := []Viewer{viewerTwo}

	if !slices.Equal(viewersOne, wantOne) {
		t.Errorf("got %s want %s", viewersOne, wantOne)
	}
	if !slices.Equal(viewersTwo, wantTwo) {
		t.Errorf("got %s want %s", viewersTwo, wantTwo)
	}
}

func TestRemoveViewer(t *testing.T) {
	channelID := models.TwitchID("channel id")
	viewerID := models.TwitchID("viewer id")
	viewer := Viewer{UserID: viewerID}

	cache := NewViewerCacheService()

	cache.AddViewer(channelID, viewer)
	cache.RemoveViewer(channelID, viewer.UserID)

	got := cache.GetViewers(channelID)
	want := []Viewer{}

	if !slices.Equal(got, want) {
		t.Errorf("got %s want %s", got, want)
	}
}

func TestUpdateViewer(t *testing.T) {
	channelID := models.TwitchID("channel id")
	viewerID := models.TwitchID("viewer id")
	viewer := Viewer{UserID: viewerID}
	image := "image"

	cache := NewViewerCacheService()

	cache.AddViewer(channelID, viewer)
	cache.UpdateViewer(channelID, viewerID, image)

	viewers := cache.GetViewers(channelID)

	if len(viewers) != 1 {
		t.Errorf("expected a singleton list but was of length %d", len(viewers))
	}
	if viewers[0].Image != image {
		t.Errorf("got %s want %s", viewers[0].Image, image)
	}
}
