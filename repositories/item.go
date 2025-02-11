package repositories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ErrItemNotFoundByName struct {
	ItemName string
}

func NewErrItemNotFoundByName(
	itemName string,
) ErrItemNotFoundByName {
	return ErrItemNotFoundByName{
		ItemName: itemName,
	}
}

func (e ErrItemNotFoundByName) Error() string {
	return "an item with the associated item name could not be found"
}

type ErrItemNotFoundById struct {
	ItemId uuid.UUID
}

func NewErrItemNotFoundById(
	itemId uuid.UUID,
) ErrItemNotFoundById {
	return ErrItemNotFoundById{
		ItemId: itemId,
	}
}

func (e ErrItemNotFoundById) Error() string {
	return "an item with the associated item id could not be found"
}

type ErrAddItemNotExist struct {
	UserId        twitch.Id
	ItemId        uuid.UUID
	TransactionId uuid.UUID
}

func NewErrAddItemNotExist(
	userId twitch.Id,
	itemId uuid.UUID,
	transactionId uuid.UUID,
) ErrAddItemNotExist {
	return ErrAddItemNotExist{
		UserId:        userId,
		ItemId:        itemId,
		TransactionId: transactionId,
	}
}

func (e ErrAddItemNotExist) Error() string {
	return "tried to add an owned item for an item that does not exist"
}

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *itemRepository {
	return &itemRepository{db: db}
}

func (repo *itemRepository) GetItemByName(channelId twitch.Id, itemName string) (item models.Item, err error) {
	result := repo.db.Joins("JOIN channel_items ON channel_items.item_id = items.item_id AND channel_items.channel_id = ? AND items.name = ?", channelId, itemName).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.Item{}, NewErrItemNotFoundByName(itemName)
	} else if result.Error != nil {
		return models.Item{}, result.Error
	}
	return
}

func (repo *itemRepository) GetItemById(itemId uuid.UUID) (item models.Item, err error) {
	result := repo.db.Where("item_id = ?", itemId).First(&item)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return models.Item{}, NewErrItemNotFoundById(itemId)
	} else if result.Error != nil {
		return models.Item{}, result.Error
	}
	return
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
	result := repo.db.Where("item_id = ?", itemId).Find(&channelItem)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return NewErrAddItemNotExist(userId, itemId, transactionId)
	} else if result.Error != nil {
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
