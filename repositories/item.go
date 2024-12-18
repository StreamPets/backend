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

func (repo *itemRepository) GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelID, itemName).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetItemByID(itemID uuid.UUID) (models.Item, error) {
	var item models.Item
	result := repo.db.Where("item_id = ?", itemID).First(&item)
	return item, result.Error
}

func (repo *itemRepository) GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error) {
	var selectedItem models.SelectedItem
	result := repo.db.Where("user_id = ? AND channel_id = ?", userID, channelID).First(&selectedItem)
	if result.Error != nil {
		return models.Item{}, result.Error
	}

	var item models.Item
	result = repo.db.Where("item_id = ?", selectedItem.ItemID).First(&item)
	return item, result.Error
}

func (repo *itemRepository) SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error {
	return repo.db.Clauses(clause.OnConflict{
		DoNothing: false,
		UpdateAll: true,
	}).Create(&models.SelectedItem{
		UserID:    userID,
		ChannelID: channelID,
		ItemID:    itemID,
	}).Error
}

func (repo *itemRepository) GetScheduledItems(channelID models.TwitchID, dayOfWeek models.DayOfWeek) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN schedules ON schedules.item_id = items.item_id AND schedules.channel_id = ? AND schedules.day_of_week = ?", channelID, dayOfWeek).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error) {
	var items []models.Item
	result := repo.db.Joins("JOIN owned_items ON owned_items.item_id = items.item_id AND owned_items.channel_id = ? AND owned_items.user_id = ?", channelID, userID).Find(&items)
	return items, result.Error
}

func (repo *itemRepository) AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error {
	var channelItem models.ChannelItem
	result := repo.db.Where("item_id = ?", itemID).Find(&channelItem)
	if result.Error != nil {
		return result.Error
	}

	result = repo.db.Create(&models.OwnedItem{
		UserID:        userID,
		ChannelID:     channelItem.ChannelID,
		ItemID:        itemID,
		TransactionID: transactionID,
	})

	return result.Error
}

func (repo *itemRepository) CheckOwnedItem(userID models.TwitchID, itemID uuid.UUID) error {
	result := repo.db.Where("user_id = ? AND item_id = ?", userID, itemID).First(&models.OwnedItem{})
	return result.Error
}
