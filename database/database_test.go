package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
	"github.com/stretchr/testify/assert"
)

func TestGetOverlayId(t *testing.T) {
	channelId := "channel id"
	overlayId := uuid.New()

	channel := models.Channel{
		ChannelId: channelId,
		OverlayId: overlayId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channel); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)
	got, err := database.GetOverlayId(channelId)

	assert.NoError(t, err)
	assert.Equal(t, overlayId, got)
}

func TestGetSelectedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	selectedItem := models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)
	got, err := database.GetSelectedItem(userId, channelId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestSetSelectedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	newItemId := uuid.New()
	newItem := models.Item{ItemId: newItemId}

	selectedItem := models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&newItem); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)

	err := database.SetSelectedItem(userId, channelId, newItemId)
	got, _ := database.GetSelectedItem(userId, channelId)

	assert.NoError(t, err)
	assert.Equal(t, newItem, got)
}

func TestDeleteSelectedItem(t *testing.T) {
	userId := "user id"
	channelId := "channel id"

	itemId := uuid.New()

	selectedItem := models.SelectedItem{
		ItemId: itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)

	err := database.DeleteSelectedItem(userId, channelId)
	assert.NoError(t, err)

	_, err = database.GetSelectedItem(userId, channelId)
	assert.Error(t, err)
}

func TestGetItemByName(t *testing.T) {
	channelId := "channel id"
	itemId := uuid.New()
	itemName := "item name"

	item := models.Item{
		ItemId: itemId,
		Name:   itemName,
	}

	channelItem := models.ChannelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)
	got, err := database.GetItemByName(channelId, itemName)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetItemById(t *testing.T) {
	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)
	got, err := database.GetItemById(itemId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetChannelsItems(t *testing.T) {
	channelId := "channel id"
	itemId := uuid.New()

	item := models.Item{
		ItemId:  itemId,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	channelItem := models.ChannelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)

	items, err := database.GetChannelsItems(channelId)
	expected := []models.Item{item}

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestGetOwnedItems(t *testing.T) {
	channelId := "channel id"
	userId := "user id"

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	owneditem := models.OwnedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&owneditem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)

	items, err := database.GetOwnedItems(channelId, userId)
	expected := []models.Item{item}

	assert.Equal(t, expected, items)
	assert.NoError(t, err)
}

func TestAddOwnedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"
	itemId := uuid.New()
	transactionId := uuid.New()

	channelItem := models.ChannelItem{
		ItemId:    itemId,
		ChannelId: channelId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	database := New(db)
	err := database.AddOwnedItem(userId, itemId, transactionId)

	assert.NoError(t, err)
}

func TestCheckOwnedItem(t *testing.T) {
	t.Run("true when user owns item", func(t *testing.T) {
		userId := "user id"
		itemId := uuid.New()

		ownedItem := models.OwnedItem{UserId: userId, ItemId: itemId}

		db := test.CreateTestDB()
		if result := db.Create(&ownedItem); result.Error != nil {
			panic(result.Error)
		}

		database := New(db)
		owned, err := database.CheckOwnedItem(userId, itemId)

		assert.NoError(t, err)
		assert.True(t, owned)
	})

	t.Run("false when item is unowned", func(t *testing.T) {
		userId := "user id"
		itemId := uuid.New()

		db := test.CreateTestDB()

		database := New(db)
		owned, err := database.CheckOwnedItem(userId, itemId)

		assert.NoError(t, err)
		assert.False(t, owned)
	})
}
