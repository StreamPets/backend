package items

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	streampets "github.com/streampets/backend"
	"github.com/stretchr/testify/assert"
)

func TestGetItemByName(t *testing.T) {
	mock.SetUp(t)

	channelId := "channel id"
	itemName := "item name"

	item := streampets.Item{Name: itemName}

	itemMock := mock.Mock[database]()
	mock.When(itemMock.ItemByName(channelId, itemName)).ThenReturn(item, nil)

	database := New(itemMock)

	got, err := database.GetItemByName(channelId, itemName)

	mock.Verify(itemMock, mock.Once()).ItemByName(channelId, itemName)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetItemById(t *testing.T) {
	mock.SetUp(t)

	itemId := uuid.New()
	item := streampets.Item{ItemId: itemId}

	itemMock := mock.Mock[database]()
	mock.When(itemMock.Item(itemId)).ThenReturn(item, nil)

	database := New(itemMock)

	got, err := database.GetItemById(itemId)

	mock.Verify(itemMock, mock.Once()).Item(itemId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetSelectedItem(t *testing.T) {
	mock.SetUp(t)

	userId := "user id"
	channelId := "channel id"
	want := streampets.Item{ItemId: uuid.New()}

	itemMock := mock.Mock[database]()
	mock.When(itemMock.SelectedItem(userId, channelId)).ThenReturn(want, nil)

	itemService := New(itemMock)

	got, err := itemService.GetSelectedItem(userId, channelId)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestSetSelectedItem(t *testing.T) {
	t.Run("item is set as selected when owned", func(t *testing.T) {
		mock.SetUp(t)

		userId := "user id"
		channelId := "channel id"
		itemId := uuid.New()

		itemMock := mock.Mock[database]()
		mock.When(itemMock.ItemOwned(userId, itemId)).ThenReturn(true, nil)

		itemService := New(itemMock)

		err := itemService.SetSelectedItem(userId, channelId, itemId)

		mock.Verify(itemMock, mock.Once()).SetSelectedItem(channelId, userId, itemId)

		assert.NoError(t, err)
	})

	t.Run("item is not set as selected when unowned", func(t *testing.T) {
		mock.SetUp(t)

		userId := "user id"
		channelId := "channel id"
		itemId := uuid.New()

		itemMock := mock.Mock[database]()
		mock.When(itemMock.ItemOwned(userId, itemId)).ThenReturn(false, nil)

		itemService := New(itemMock)

		mock.Verify(itemMock, mock.Never()).SetSelectedItem(channelId, userId, itemId)

		err := itemService.SetSelectedItem(userId, channelId, itemId)
		if assert.Error(t, err) {
			assert.Equal(t, ErrSelectUnownedItem, err)
		}
	})
}

func TestGetChannelsItems(t *testing.T) {
	mock.SetUp(t)

	channelId := "channel id"
	expected := []streampets.Item{{}}

	itemMock := mock.Mock[database]()
	mock.When(itemMock.ItemsByChannelId(channelId)).ThenReturn(expected, nil)

	itemService := New(itemMock)

	items, err := itemService.GetChannelsItems(channelId)

	mock.Verify(itemMock, mock.Once()).ItemsByChannelId(channelId)

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestGetOwnedItems(t *testing.T) {
	mock.SetUp(t)

	channelId := "channel id"
	userId := "user id"
	expected := []streampets.Item{{}}

	itemMock := mock.Mock[database]()

	mock.When(itemMock.ItemsByUserId(channelId, userId)).ThenReturn(expected, nil)

	itemService := New(itemMock)

	items, err := itemService.GetOwnedItems(channelId, userId)

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestAddOwnedItem(t *testing.T) {
	mock.SetUp(t)

	userId := "user id"
	itemId := uuid.New()
	transactionId := uuid.New()

	itemMock := mock.Mock[database]()
	mock.When(itemMock.CreateOwnedItem(userId, itemId, transactionId)).ThenReturn(nil)

	itemService := New(itemMock)

	err := itemService.AddOwnedItem(userId, itemId, transactionId)

	assert.NoError(t, err)
}
