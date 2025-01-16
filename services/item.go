package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

var ErrSelectUnknownItem = errors.New("viewer tried to select an item they do not own")

type ItemRepository interface {
	GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error)
	GetItemById(itemId uuid.UUID) (models.Item, error)

	GetSelectedItem(viewerId, channelId models.TwitchId) (models.Item, error)
	SetSelectedItem(viewerId, channelId models.TwitchId, itemId uuid.UUID) error

	GetChannelsItems(channelId models.TwitchId) ([]models.Item, error)

	GetOwnedItems(channelId, viewerId models.TwitchId) ([]models.Item, error)
	AddOwnedItem(viewerId models.TwitchId, itemId, transactionId uuid.UUID) error
	CheckOwnedItem(viewerId models.TwitchId, itemId uuid.UUID) (bool, error)
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

func (s *ItemService) GetSelectedItem(viewerId, channelId models.TwitchId) (models.Item, error) {
	return s.itemRepo.GetSelectedItem(viewerId, channelId)
}

func (s *ItemService) SetSelectedItem(viewerId, channelId models.TwitchId, itemId uuid.UUID) error {
	owned, err := s.itemRepo.CheckOwnedItem(viewerId, itemId)
	if err != nil {
		return err
	}
	if !owned {
		return ErrSelectUnknownItem
	}

	return s.itemRepo.SetSelectedItem(channelId, viewerId, itemId)
}

func (s *ItemService) GetChannelsItems(channelId models.TwitchId) ([]models.Item, error) {
	return s.itemRepo.GetChannelsItems(channelId)
}

func (s *ItemService) GetOwnedItems(channelId, viewerId models.TwitchId) ([]models.Item, error) {
	return s.itemRepo.GetOwnedItems(channelId, viewerId)
}

func (s *ItemService) AddOwnedItem(viewerId models.TwitchId, itemId, transactionId uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(viewerId, itemId, transactionId)
}
