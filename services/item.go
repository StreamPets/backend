package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
)

var ErrSelectUnownedItem = errors.New("user tried to select an item they do not own")

type ItemRepository interface {
	GetItemByName(channelId twitch.Id, itemName string) (models.Item, error)
	GetItemById(itemId uuid.UUID) (models.Item, error)

	GetSelectedItem(userId, channelId twitch.Id) (models.Item, error)
	SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error
	DeleteSelectedItem(userId, channelId twitch.Id) error

	GetChannelsItems(channelId twitch.Id) ([]models.Item, error)

	GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error)
	AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error
	CheckOwnedItem(userId twitch.Id, itemId uuid.UUID) (bool, error)

	GetDefaultItem(channelId twitch.Id) (models.Item, error)
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

func (s *ItemService) GetItemByName(channelId twitch.Id, itemName string) (models.Item, error) {
	return s.itemRepo.GetItemByName(channelId, itemName)
}

func (s *ItemService) GetItemById(itemId uuid.UUID) (models.Item, error) {
	return s.itemRepo.GetItemById(itemId)
}

func (s *ItemService) GetSelectedItem(userId, channelId twitch.Id) (models.Item, error) {
	item, err := s.itemRepo.GetSelectedItem(userId, channelId)
	if err == gorm.ErrRecordNotFound {
		return s.itemRepo.GetDefaultItem(channelId)
	}
	if err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (s *ItemService) SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error {
	if owned, err := s.itemRepo.CheckOwnedItem(userId, itemId); err != nil {
		return err
	} else if owned {
		return s.itemRepo.SetSelectedItem(channelId, userId, itemId)
	}

	if defaultItem, err := s.itemRepo.GetDefaultItem(channelId); err != nil {
		return err
	} else if defaultItem.ItemId != itemId {
		return ErrSelectUnownedItem
	}

	return s.itemRepo.DeleteSelectedItem(userId, channelId)
}

func (s *ItemService) GetChannelsItems(channelId twitch.Id) ([]models.Item, error) {
	return s.itemRepo.GetChannelsItems(channelId)
}

func (s *ItemService) GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error) {
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

func (s *ItemService) AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(userId, itemId, transactionId)
}
