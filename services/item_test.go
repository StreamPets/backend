package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestGetItemByName(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	itemName := "item name"

	item := models.Item{Name: itemName}

	itemMock := mock.Mock[ItemRepository]()
	mock.When(itemMock.GetItemByName(channelId, itemName)).ThenReturn(item, nil)

	database := NewItemService(itemMock)

	got, err := database.GetItemByName(channelId, itemName)

	mock.Verify(itemMock, mock.Once()).GetItemByName(channelId, itemName)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetItemById(t *testing.T) {
	mock.SetUp(t)

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	itemMock := mock.Mock[ItemRepository]()
	mock.When(itemMock.GetItemById(itemId)).ThenReturn(item, nil)

	database := NewItemService(itemMock)

	got, err := database.GetItemById(itemId)

	mock.Verify(itemMock, mock.Once()).GetItemById(itemId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetSelectedItem(t *testing.T) {
	mock.SetUp(t)

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	want := models.Item{ItemId: uuid.New()}

	itemMock := mock.Mock[ItemRepository]()
	mock.When(itemMock.GetSelectedItem(userId, channelId)).ThenReturn(want, nil)

	itemService := NewItemService(itemMock)

	got, err := itemService.GetSelectedItem(userId, channelId)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestSetSelectedItem(t *testing.T) {
	t.Run("item is set as selected when owned", func(t *testing.T) {
		mock.SetUp(t)

		userId := models.TwitchId("user id")
		channelId := models.TwitchId("channel id")
		itemId := uuid.New()

		itemMock := mock.Mock[ItemRepository]()
		mock.When(itemMock.CheckOwnedItem(userId, itemId)).ThenReturn(true, nil)

		itemService := NewItemService(itemMock)

		err := itemService.SetSelectedItem(userId, channelId, itemId)

		mock.Verify(itemMock, mock.Once()).SetSelectedItem(channelId, userId, itemId)

		assert.NoError(t, err)
	})

	t.Run("item is not set as selected when unowned", func(t *testing.T) {
		mock.SetUp(t)

		userId := models.TwitchId("user id")
		channelId := models.TwitchId("channel id")
		itemId := uuid.New()

		itemMock := mock.Mock[ItemRepository]()
		mock.When(itemMock.CheckOwnedItem(userId, itemId)).ThenReturn(false, nil)

		itemService := NewItemService(itemMock)

		mock.Verify(itemMock, mock.Never()).SetSelectedItem(channelId, userId, itemId)

		err := itemService.SetSelectedItem(userId, channelId, itemId)
		if assert.Error(t, err) {
			assert.Equal(t, ErrSelectUnownedItem, err)
		}
	})
}

func TestGetChannelsItems(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	expected := []models.Item{{}}

	itemMock := mock.Mock[ItemRepository]()
	mock.When(itemMock.GetChannelsItems(channelId)).ThenReturn(expected, nil)

	itemService := NewItemService(itemMock)

	items, err := itemService.GetChannelsItems(channelId)

	mock.Verify(itemMock, mock.Once()).GetChannelsItems(channelId)

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestGetOwnedItems(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	expected := []models.Item{{}}

	itemMock := mock.Mock[ItemRepository]()

	mock.When(itemMock.GetOwnedItems(channelId, userId)).ThenReturn(expected, nil)

	itemService := NewItemService(itemMock)

	items, err := itemService.GetOwnedItems(channelId, userId)

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestAddOwnedItem(t *testing.T) {
	mock.SetUp(t)

	userId := models.TwitchId("user id")
	itemId := uuid.New()
	transactionId := uuid.New()

	itemMock := mock.Mock[ItemRepository]()
	mock.When(itemMock.AddOwnedItem(userId, itemId, transactionId)).ThenReturn(nil)

	itemService := NewItemService(itemMock)

	err := itemService.AddOwnedItem(userId, itemId, transactionId)

	assert.NoError(t, err)
}
