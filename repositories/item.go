package repositories

import (
	"os"

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

func (repo *itemRepository) GetItemByName(channelId models.UserId, itemName string) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetItemById(itemId uuid.UUID) (models.Item, error) {
	var item models.Item
	result := repo.db.Where("item_id = ?", itemId).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetSelectedItem(viewerId, channelId models.UserId) (models.Item, error) {
	var selectedItem models.SelectedItem
	result := repo.db.Where("viewer_id = ? AND channel_id = ?", viewerId, channelId).First(&selectedItem)

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

func (repo *itemRepository) SetSelectedItem(viewerId, channelId models.UserId, itemId uuid.UUID) error {
	return repo.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&models.SelectedItem{
		ViewerId:  viewerId,
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}

func (repo *itemRepository) GetChannelsItems(channelId models.UserId) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ?", channelId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) GetOwnedItems(channelId, viewerId models.UserId) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.viewer_id = ?", channelId, viewerId).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) AddOwnedItem(viewerId models.UserId, itemId, transactionId uuid.UUID) error {
	var channelItem models.ChannelItem
	result := repo.db.Where("item_id = ?", itemId).Find(&channelItem)
	if result.Error != nil {
		return result.Error
	}

	if os.Getenv("ENVIRONMENT") == "DEVELOPMENT" {
		transactionId = uuid.New()
	}

	result = repo.db.Create(&models.OwnedItem{
		ViewerId:      viewerId,
		ChannelId:     channelItem.ChannelId,
		ItemId:        itemId,
		TransactionId: transactionId,
	})

	return result.Error
}

func (repo *itemRepository) CheckOwnedItem(viewerId models.UserId, itemId uuid.UUID) (bool, error) {
	result := repo.db.Where("viewer_id = ? AND item_id = ?", viewerId, itemId).First(&models.OwnedItem{})
	if result.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	if result.Error != nil {
		return false, result.Error
	}

	return true, nil
}
