package services_test

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

func TestGetViewer(t *testing.T) {
	mock.SetUp(t)

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	channelName := "channel name"
	username := "username"
	image := "image"

	item := models.Item{Image: image}

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	mock.When(itemMock.GetSelectedItem(userID, channelID)).ThenReturn(item, nil)
	mock.When(twitchMock.GetUserID(channelName)).ThenReturn(channelID, nil)

	databaseService := services.NewDatabaseService(itemMock, twitchMock)

	viewer, err := databaseService.GetViewer(userID, channelName, username)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	expected := services.Viewer{
		UserID:   userID,
		Username: username,
		Image:    image,
	}

	mock.Verify(itemMock, mock.Once()).GetSelectedItem(userID, channelID)
	if viewer != expected {
		t.Errorf("expected %s got %s", expected, viewer)
	}
}

func TestGetItemByName(t *testing.T) {
	mock.SetUp(t)

	channelID := models.TwitchID("channel id")
	itemName := "item name"

	item := models.Item{Name: itemName}

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	mock.When(itemMock.GetItemByName(channelID, itemName)).ThenReturn(item, nil)

	database := services.NewDatabaseService(itemMock, twitchMock)

	got, err := database.GetItemByName(channelID, itemName)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	if got != item {
		t.Errorf("expected %s got %s", item, got)
	}

	mock.Verify(itemMock, mock.Once()).GetItemByName(channelID, itemName)
}

func TestGetItemByID(t *testing.T) {
	mock.SetUp(t)

	itemID := uuid.New()
	item := models.Item{ItemID: itemID}

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	mock.When(itemMock.GetItemByID(itemID)).ThenReturn(item, nil)

	database := services.NewDatabaseService(itemMock, twitchMock)

	got, err := database.GetItemByID(itemID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	if got != item {
		t.Errorf("expected %s got %s", item, got)
	}

	mock.Verify(itemMock, mock.Once()).GetItemByID(itemID)
}

func TestSetSelectedItem(t *testing.T) {
	mock.SetUp(t)

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	itemID := uuid.New()

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	database := services.NewDatabaseService(itemMock, twitchMock)

	err := database.SetSelectedItem(userID, channelID, itemID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	mock.Verify(itemMock, mock.Once()).SetSelectedItem(channelID, userID, itemID)
}

func TestGetTodaysItems(t *testing.T) {
	mock.SetUp(t)

	currentTime := time.Now()
	dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())
	channelID := models.TwitchID("channel id")

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	expected := []models.Item{{}}

	mock.When(itemMock.GetScheduledItems(channelID, dayOfWeek)).ThenReturn(expected, nil)

	databaseService := services.NewDatabaseService(itemMock, twitchMock)

	items, err := databaseService.GetTodaysItems(channelID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	mock.Verify(itemMock, mock.Once()).GetScheduledItems(channelID, dayOfWeek)

	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestGetOwnedItems(t *testing.T) {
	mock.SetUp(t)

	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	expected := []models.Item{{}}

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	mock.When(itemMock.GetOwnedItems(channelID, userID)).ThenReturn(expected, nil)

	databaseService := services.NewDatabaseService(itemMock, twitchMock)

	items, err := databaseService.GetOwnedItems(channelID, userID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}

	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestAddOwnedItem(t *testing.T) {
	mock.SetUp(t)

	userID := models.TwitchID("user id")
	itemID := uuid.New()
	transactionID := uuid.New()

	itemMock := mock.Mock[repositories.ItemRepository]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()

	mock.When(itemMock.AddOwnedItem(userID, itemID, transactionID)).ThenReturn(nil)

	databaseService := services.NewDatabaseService(itemMock, twitchMock)

	err := databaseService.AddOwnedItem(userID, itemID, transactionID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}
}
