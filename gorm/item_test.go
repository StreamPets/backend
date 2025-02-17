package gorm

import (
	"testing"

	"github.com/google/uuid"
	streampets "github.com/streampets/backend"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestItem(t *testing.T) {
	itemId := uuid.New()
	item := streampets.Item{ItemId: itemId}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&item)
	assert.NoError(t, result.Error)

	repository := NewItemRepository(db)
	got, err := repository.Item(itemId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestItemByName(t *testing.T) {
	channelId := "channel id"
	itemId := uuid.New()
	itemName := "item name"

	item := streampets.Item{
		ItemId: itemId,
		Name:   itemName,
	}

	channelItem := channelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&item)
	assert.NoError(t, result.Error)

	result = db.Create(&channelItem)
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)
	got, err := database.ItemByName(channelId, itemName)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestGetSelectedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"
	itemId := uuid.New()

	item := streampets.Item{
		ItemId: itemId,
	}

	selectedItem := selectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&item)
	assert.NoError(t, result.Error)

	result = db.Create(&selectedItem)
	assert.NoError(t, result.Error)

	repo := NewItemRepository(db)
	got, err := repo.SelectedItem(channelId, userId)

	assert.NoError(t, err)
	assert.Equal(t, item, got)
}

func TestSetSelectedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"

	itemId := uuid.New()
	newItemId := uuid.New()
	newItem := streampets.Item{ItemId: newItemId}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&streampets.Item{
		ItemId: itemId,
	})
	assert.NoError(t, result.Error)

	result = db.Create(&newItem)
	assert.NoError(t, result.Error)

	result = db.Create(&selectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	})
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)

	err = database.SetSelectedItem(channelId, userId, newItemId)
	assert.NoError(t, err)

	got, err := database.SelectedItem(channelId, userId)
	assert.NoError(t, err)

	assert.Equal(t, newItem, got)
}

func TestDeleteSelectedItem(t *testing.T) {
	userId := "user id"
	channelId := "channel id"

	itemId := uuid.New()

	selectedItem := selectedItem{
		ItemId: itemId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&selectedItem)
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)

	err = database.DeleteSelectedItem(userId, channelId)
	assert.NoError(t, err)

	_, err = database.SelectedItem(userId, channelId)
	if assert.Error(t, err) {
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	}
}

func TestGetChannelsItems(t *testing.T) {
	channelId := "channel id"
	itemId := uuid.New()

	item := streampets.Item{
		ItemId:  itemId,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	channelItem := &channelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&item)
	assert.NoError(t, result.Error)

	result = db.Create(&channelItem)
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)

	items, err := database.ItemsByChannelId(channelId)
	expected := []streampets.Item{item}

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func TestGetOwnedItems(t *testing.T) {
	channelId := "channel id"
	userId := "user id"

	itemId := uuid.New()
	item := streampets.Item{ItemId: itemId}

	owneditem := ownedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&item)
	assert.NoError(t, result.Error)

	result = db.Create(&owneditem)
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)

	items, err := database.ItemsByUserId(channelId, userId)
	expected := []streampets.Item{item}

	assert.Equal(t, expected, items)
	assert.NoError(t, err)
}

func TestAddOwnedItem(t *testing.T) {
	channelId := "channel id"
	userId := "user id"
	itemId := uuid.New()
	transactionId := uuid.New()

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&channelItem{
		ItemId:    itemId,
		ChannelId: channelId,
	})
	assert.NoError(t, result.Error)

	database := NewItemRepository(db)
	err = database.CreateOwnedItem(userId, itemId, transactionId)

	assert.NoError(t, err)
}

func TestCheckOwnedItem(t *testing.T) {
	t.Run("true when user owns item", func(t *testing.T) {
		userId := "user id"
		itemId := uuid.New()

		ownedItem := ownedItem{UserId: userId, ItemId: itemId}

		db, err := CreateTestDB()
		assert.NoError(t, err)

		result := db.Create(&ownedItem)
		assert.NoError(t, result.Error)

		database := NewItemRepository(db)
		owned, err := database.ItemOwned(userId, itemId)

		assert.NoError(t, err)
		assert.True(t, owned)
	})

	t.Run("false when item is unowned", func(t *testing.T) {
		userId := "user id"
		itemId := uuid.New()

		db, err := CreateTestDB()
		assert.NoError(t, err)

		database := NewItemRepository(db)
		owned, err := database.ItemOwned(userId, itemId)

		assert.NoError(t, err)
		assert.False(t, owned)
	})
}
