package repositories

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *itemRepository {
	return &itemRepository{db: db}
}

func (repo *itemRepository) GetItemByName(channelId twitch.Id, itemName string) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetItemById(itemId uuid.UUID) (models.Item, error) {
	var item models.Item
	result := repo.db.Where("item_id = ?", itemId).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetSelectedItem(userId, channelId twitch.Id) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins(`JOIN selected_items ON selected_items.item_id = items.item_id AND selected_items.user_id = ? AND selected_items.channel_id = ?`, userId, channelId).First(&item)
	return item, result.Error
}

func (repo *itemRepository) SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error {
	return repo.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}

func (repo *itemRepository) DeleteSelectedItem(userId, channelId twitch.Id) error {
	selectedItem := models.SelectedItem{UserId: userId, ChannelId: channelId}
	return repo.db.Delete(&selectedItem).Error
}

func (repo *itemRepository) GetChannelsItems(channelId twitch.Id) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ?", channelId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.user_id = ?", channelId, userId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error {
	var channelItem models.ChannelItem
	if result := repo.db.Where("item_id = ?", itemId).Find(&channelItem); result.Error != nil {
		return result.Error
	}

	return repo.db.Create(&models.OwnedItem{
		UserId:        userId,
		ChannelId:     channelItem.ChannelId,
		ItemId:        itemId,
		TransactionId: transactionId,
	}).Error
}

func (repo *itemRepository) CheckOwnedItem(userId twitch.Id, itemId uuid.UUID) (bool, error) {
	result := repo.db.Where("user_id = ? AND item_id = ?", userId, itemId).First(&models.OwnedItem{})
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	} else if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}

func (repo *itemRepository) GetDefaultItem(channelId twitch.Id) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN default_channel_items ON default_channel_items.item_id = items.item_id AND default_channel_items.channel_id = ?", channelId).First(&item)
	return item, result.Error
}
