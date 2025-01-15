package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

var ErrSelectUnknownItem = errors.New("user tried to select an item they do not own")

type ItemRepository interface {
	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
	GetItemByID(itemID uuid.UUID) (models.Item, error)

	GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error

	GetScheduledItems(channelID models.TwitchID, dayOfWeek models.DayOfWeek) ([]models.Item, error)
	GetChannelItems(channelID models.TwitchID) ([]models.Item, error)

	GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error)
	AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error
	CheckOwnedItem(userID models.TwitchID, itemID uuid.UUID) (bool, error)
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

func (s *ItemService) GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error) {
	return s.itemRepo.GetItemByName(channelID, itemName)
}

func (s *ItemService) GetItemByID(itemID uuid.UUID) (models.Item, error) {
	return s.itemRepo.GetItemByID(itemID)
}

func (s *ItemService) GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error) {
	return s.itemRepo.GetSelectedItem(userID, channelID)
}

func (s *ItemService) SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error {
	owned, err := s.itemRepo.CheckOwnedItem(userID, itemID)
	if err != nil {
		return err
	}
	if !owned {
		return ErrSelectUnknownItem
	}

	return s.itemRepo.SetSelectedItem(channelID, userID, itemID)
}

func (s *ItemService) GetTodaysItems(channelID models.TwitchID) ([]models.Item, error) {
	// currentTime := time.Now()
	// dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())

	// return s.itemRepo.GetScheduledItems(channelID, dayOfWeek)

	return s.itemRepo.GetChannelItems(channelID)
}

func (s *ItemService) GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error) {
	return s.itemRepo.GetOwnedItems(channelID, userID)
}

func (s *ItemService) AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(userID, itemID, transactionID)
}
