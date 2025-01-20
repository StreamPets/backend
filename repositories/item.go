package repositories

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *itemRepository {
	return &itemRepository{db: db}
}

func (repo *itemRepository) GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetItemById(itemId uuid.UUID) (models.Item, error) {
	var item models.Item
	result := repo.db.Where("item_id = ?", itemId).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error) {
	var selectedItem models.SelectedItem
	result := repo.db.Where("user_id = ? AND channel_id = ?", userId, channelId).First(&selectedItem)

	if result.Error == nil {
		var item models.Item
		result = repo.db.Where("item_id = ?", selectedItem.ItemId).First(&item)
		return item, result.Error
	}

	if result.Error != gorm.ErrRecordNotFound {
		return models.Item{}, result.Error
	}

	var defaultChannelItem models.DefaultChannelItem
	result = repo.db.Where("channel_id = ?", channelId).First(&defaultChannelItem)
	if result.Error != nil {
		return models.Item{}, result.Error
	}

	var item models.Item
	result = repo.db.Where("item_id = ?", defaultChannelItem.ItemId).First(&item)
	return item, result.Error
}

func (repo *itemRepository) SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error {
	return repo.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}

func (repo *itemRepository) GetChannelsItems(channelId models.TwitchId) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ?", channelId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) GetOwnedItems(channelId, userId models.TwitchId) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.user_id = ?", channelId, userId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) AddOwnedItem(userId models.TwitchId, itemId, transactionId uuid.UUID) error {
	var channelItem models.ChannelItem
	result := repo.db.Where("item_id = ?", itemId).Find(&channelItem)
	if result.Error != nil {
		return result.Error
	}

	result = repo.db.Create(&models.OwnedItem{
		UserId:        userId,
		ChannelId:     channelItem.ChannelId,
		ItemId:        itemId,
		TransactionId: transactionId,
	})

	return result.Error
}

func (repo *itemRepository) CheckOwnedItem(userId models.TwitchId, itemId uuid.UUID) (bool, error) {
	result := repo.db.Where("user_id = ? AND item_id = ?", userId, itemId).First(&models.OwnedItem{})
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}
