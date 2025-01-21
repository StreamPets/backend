package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
)

var ErrSelectUnownedItem = errors.New("user tried to select an item they do not own")

type ItemRepository interface {
	GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error)
	GetItemById(itemId uuid.UUID) (models.Item, error)

	GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error)
	SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error
	DeleteSelectedItem(userId, channelId models.TwitchId) error

	GetChannelsItems(channelId models.TwitchId) ([]models.Item, error)

	GetOwnedItems(channelId, userId models.TwitchId) ([]models.Item, error)
	AddOwnedItem(userId models.TwitchId, itemId, transactionId uuid.UUID) error
	CheckOwnedItem(userId models.TwitchId, itemId uuid.UUID) (bool, error)

	GetDefaultItem(channelId models.TwitchId) (models.Item, error)
}

type ItemService struct {
	itemRepo ItemRepository
}

func NewItemService(
	itemRepo ItemRepository,
) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
	}
}

func (s *ItemService) GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error) {
	return s.itemRepo.GetItemByName(channelId, itemName)
}

func (s *ItemService) GetItemById(itemId uuid.UUID) (models.Item, error) {
	return s.itemRepo.GetItemById(itemId)
}

func (s *ItemService) GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error) {
	item, err := s.itemRepo.GetSelectedItem(userId, channelId)
	if err == gorm.ErrRecordNotFound {
		return s.itemRepo.GetDefaultItem(channelId)
	}
	if err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (s *ItemService) SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error {
	owned, err := s.itemRepo.CheckOwnedItem(userId, itemId)
	if err != nil {
		return err
	}

	if owned {
		return s.itemRepo.SetSelectedItem(channelId, userId, itemId)
	}

	defaultItem, err := s.itemRepo.GetDefaultItem(channelId)
	if err != nil {
		return err
	}

	if defaultItem.ItemId == itemId {
		return s.itemRepo.DeleteSelectedItem(userId, channelId)
	}

	return ErrSelectUnownedItem
}

func (s *ItemService) GetChannelsItems(channelId models.TwitchId) ([]models.Item, error) {
	return s.itemRepo.GetChannelsItems(channelId)
}

func (s *ItemService) GetOwnedItems(channelId, userId models.TwitchId) ([]models.Item, error) {
	ownedItems, err := s.itemRepo.GetOwnedItems(channelId, userId)
	if err != nil {
		return []models.Item{}, err
	}

	items := map[models.Item]bool{}
	for _, ownedItem := range ownedItems {
		items[ownedItem] = true
	}

	defaultItem, err := s.itemRepo.GetDefaultItem(channelId)
	if err != nil {
		return []models.Item{}, err
	}
	items[defaultItem] = true

	result := []models.Item{}
	for item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (s *ItemService) AddOwnedItem(userId models.TwitchId, itemId, transactionId uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(userId, itemId, transactionId)
}
