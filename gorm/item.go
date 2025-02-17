package gorm

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	streampets "github.com/streampets/backend"
)

type ItemRepo struct {
	db *DB
}

func NewItemRepository(db *DB) *ItemRepo {
	return &ItemRepo{
		db: db,
	}
}

func (r *ItemRepo) Item(itemId uuid.UUID) (item streampets.Item, err error) {
	result := r.db.Where("item_id = ?", itemId).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return streampets.Item{}, ErrItemNotFound
	} else if result.Error != nil {
		return streampets.Item{}, result.Error
	}
	return
}

func (r *ItemRepo) ItemByName(channelId string, itemName string) (item streampets.Item, err error) {
	result := r.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return streampets.Item{}, ErrItemNotFoundByName
	} else if result.Error != nil {
		return streampets.Item{}, result.Error
	}
	return
}

func (r *ItemRepo) ItemsByChannelId(channelId string) ([]streampets.Item, error) {
	var items []streampets.Item
	result := r.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ?", channelId).Find(&items)
	return items, result.Error
}

func (r *ItemRepo) ItemsByUserId(channelId, userId string) (items []streampets.Item, err error) {
	result := r.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.user_id = ?", channelId, userId).Find(&items)
	return items, result.Error
}

func (r *ItemRepo) DefaultItem(channelId string) (streampets.Item, error) {
	var item streampets.Item
	result := r.db.Joins("JOIN default_channel_items ON default_channel_items.item_id = items.item_id AND default_channel_items.channel_id = ?", channelId).First(&item)
	return item, result.Error
}

func (r *ItemRepo) CreateOwnedItem(userId string, itemId, transactionId uuid.UUID) error {
	var channelItem channelItem
	result := r.db.Where("item_id = ?", itemId).Find(&channelItem)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return ErrAddItemNotExist
	} else if result.Error != nil {
		return result.Error
	}

	return r.db.Create(&ownedItem{
		UserId:        userId,
		ChannelId:     channelItem.ChannelId,
		ItemId:        itemId,
		TransactionId: transactionId,
	}).Error
}

func (r *ItemRepo) ItemOwned(userId string, itemId uuid.UUID) (bool, error) {
	result := r.db.Where("user_id = ? AND item_id = ?", userId, itemId).First(&ownedItem{})
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (r *ItemRepo) SelectedItem(channelId, userId string) (streampets.Item, error) {
	var item streampets.Item
	result := r.db.Joins(`JOIN selected_items ON selected_items.item_id = items.item_id AND selected_items.user_id = ? AND selected_items.channel_id = ?`, userId, channelId).First(&item)
	return item, result.Error
}

func (r *ItemRepo) SetSelectedItem(channelId, userId string, itemId uuid.UUID) error {
	return r.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&selectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}

func (r *ItemRepo) DeleteSelectedItem(channelId, userId string) error {
	selectedItem := selectedItem{UserId: userId, ChannelId: channelId}
	return r.db.Delete(&selectedItem).Error
}
