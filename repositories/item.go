package repositories

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ItemRepository interface {
	GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error
	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
}

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{db: db}
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

func (repo *itemRepository) GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error) {
	var item models.Item
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelID, itemName).First(&item)
	return item, result.Error
}
