package database

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	db *gorm.DB
}

func New(db *gorm.DB) *DB {
	return &DB{db: db}
}

func (db *DB) GetOverlayId(channelId twitch.Id) (uuid.UUID, error) {
	var channel models.Channel

	if result := db.db.Where("channel_id = ?", channelId).First(&channel); result.Error == gorm.ErrRecordNotFound {
		return uuid.UUID{}, ErrNoOverlayId
	} else if result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}

func (db *DB) GetItemByName(channelId twitch.Id, itemName string) (item models.Item, err error) {
	result := db.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.Item{}, NewErrItemNotFoundByName(itemName)
	} else if result.Error != nil {
		return models.Item{}, result.Error
	}
	return
}

func (db *DB) GetItemById(itemId uuid.UUID) (item models.Item, err error) {
	result := db.db.Where("item_id = ?", itemId).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.Item{}, NewErrItemNotFoundById(itemId)
	} else if result.Error != nil {
		return models.Item{}, result.Error
	}
	return
}

func (db *DB) GetSelectedItem(userId, channelId twitch.Id) (models.Item, error) {
	var item models.Item
	result := db.db.Joins(`JOIN selected_items ON selected_items.item_id = items.item_id AND selected_items.user_id = ? AND selected_items.channel_id = ?`, userId, channelId).First(&item)
	return item, result.Error
}

func (db *DB) SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error {
	return db.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}

func (db *DB) DeleteSelectedItem(userId, channelId twitch.Id) error {
	selectedItem := models.SelectedItem{UserId: userId, ChannelId: channelId}
	return db.db.Delete(&selectedItem).Error
}

func (db *DB) GetChannelsItems(channelId twitch.Id) ([]models.Item, error) {
	var items []models.Item
	result := db.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ?", channelId).Find(&items)
	return items, result.Error
}

func (db *DB) GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error) {
	var items []models.Item
	result := db.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.user_id = ?", channelId, userId).Find(&items)
	return items, result.Error
}

func (db *DB) AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error {
	var channelItem models.ChannelItem
	result := db.db.Where("item_id = ?", itemId).Find(&channelItem)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return NewErrAddItemNotExist(userId, itemId, transactionId)
	} else if result.Error != nil {
		return result.Error
	}

	return db.db.Create(&models.OwnedItem{
		UserId:        userId,
		ChannelId:     channelItem.ChannelId,
		ItemId:        itemId,
		TransactionId: transactionId,
	}).Error
}

func (db *DB) CheckOwnedItem(userId twitch.Id, itemId uuid.UUID) (bool, error) {
	result := db.db.Where("user_id = ? AND item_id = ?", userId, itemId).First(&models.OwnedItem{})
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (db *DB) GetDefaultItem(channelId twitch.Id) (models.Item, error) {
	var item models.Item
	result := db.db.Joins("JOIN default_channel_items ON default_channel_items.item_id = items.item_id AND default_channel_items.channel_id = ?", channelId).First(&item)
	return item, result.Error
}
