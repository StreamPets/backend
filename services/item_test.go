package services

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestGetItemByName(t *testing.T) {
	mock.SetUp(t)

	channelID := models.TwitchID("channel id")
	itemName := "item name"

	item := models.Item{Name: itemName}

	itemMock := mock.Mock[ItemRepository]()

	mock.When(itemMock.GetItemByName(channelID, itemName)).ThenReturn(item, nil)

	database := NewItemService(itemMock)

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

	itemMock := mock.Mock[ItemRepository]()

	mock.When(itemMock.GetItemByID(itemID)).ThenReturn(item, nil)

	database := NewItemService(itemMock)

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
	t.Run("item is set as selected when owned", func(t *testing.T) {
		mock.SetUp(t)

		userID := models.TwitchID("user id")
		channelID := models.TwitchID("channel id")
		itemID := uuid.New()

		itemMock := mock.Mock[ItemRepository]()

		database := NewItemService(itemMock)

		mock.When(itemMock.CheckOwnedItem(userID, itemID)).ThenReturn(true, nil)

		err := database.SetSelectedItem(userID, channelID, itemID)
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		mock.Verify(itemMock, mock.Once()).SetSelectedItem(channelID, userID, itemID)
	})

	t.Run("item is not set as selected when unowned", func(t *testing.T) {
		mock.SetUp(t)

		userID := models.TwitchID("user id")
		channelID := models.TwitchID("channel id")
		itemID := uuid.New()

		itemMock := mock.Mock[ItemRepository]()

		mock.When(itemMock.CheckOwnedItem(userID, itemID)).ThenReturn(false, nil)

		database := NewItemService(itemMock)

		err := database.SetSelectedItem(userID, channelID, itemID)
		if err == nil {
			t.Errorf("expected an error but did not receive one")
		}
		if err != ErrSelectUnknownItem {
			t.Errorf("expected %s got %s", ErrSelectUnknownItem.Error(), err.Error())
		}

		mock.Verify(itemMock, mock.Never()).SetSelectedItem(channelID, userID, itemID)
	})
}

func TestGetTodaysItems(t *testing.T) {
	mock.SetUp(t)

	currentTime := time.Now()
	dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())
	channelID := models.TwitchID("channel id")

	itemMock := mock.Mock[ItemRepository]()

	expected := []models.Item{{}}

	mock.When(itemMock.GetScheduledItems(channelID, dayOfWeek)).ThenReturn(expected, nil)

	itemService := NewItemService(itemMock)

	items, err := itemService.GetTodaysItems(channelID)
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

	itemMock := mock.Mock[ItemRepository]()

	mock.When(itemMock.GetOwnedItems(channelID, userID)).ThenReturn(expected, nil)

	itemService := NewItemService(itemMock)

	items, err := itemService.GetOwnedItems(channelID, userID)
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

	itemMock := mock.Mock[ItemRepository]()

	mock.When(itemMock.AddOwnedItem(userID, itemID, transactionID)).ThenReturn(nil)

	itemService := NewItemService(itemMock)

	err := itemService.AddOwnedItem(userID, itemID, transactionID)
	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}
}
