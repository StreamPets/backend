package items

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
)

type database interface {
	GetItemByName(channelId twitch.Id, itemName string) (models.Item, error)
	GetItemById(itemId uuid.UUID) (models.Item, error)

	GetSelectedItem(userId, channelId twitch.Id) (models.Item, error)
	SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error
	DeleteSelectedItem(userId, channelId twitch.Id) error

	GetChannelsItems(channelId twitch.Id) ([]models.Item, error)
	GetDefaultItem(channelId twitch.Id) (models.Item, error)

	GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error)
	AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error
	CheckOwnedItem(userId twitch.Id, itemId uuid.UUID) (bool, error)
}

type ItemService struct {
	db database
}

func New(db database) *ItemService {
	return &ItemService{db: db}
}

func (s *ItemService) GetItemByName(channelId twitch.Id, itemName string) (models.Item, error) {
	return s.db.GetItemByName(channelId, itemName)
}

func (s *ItemService) GetItemById(itemId uuid.UUID) (models.Item, error) {
	return s.db.GetItemById(itemId)
}

func (s *ItemService) GetSelectedItem(userId, channelId twitch.Id) (models.Item, error) {
	item, err := s.db.GetSelectedItem(userId, channelId)
	if err == gorm.ErrRecordNotFound {
		return s.db.GetDefaultItem(channelId)
	} else if err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (s *ItemService) SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error {
	if owned, err := s.db.CheckOwnedItem(userId, itemId); err != nil {
		return err
	} else if owned {
		return s.db.SetSelectedItem(channelId, userId, itemId)
	}

	if defaultItem, err := s.db.GetDefaultItem(channelId); err != nil {
		return err
	} else if defaultItem.ItemId != itemId {
		return NewErrSelectUnownedItem(userId, channelId, itemId)
	}

	return s.db.DeleteSelectedItem(userId, channelId)
}

func (s *ItemService) GetChannelsItems(channelId twitch.Id) ([]models.Item, error) {
	return s.db.GetChannelsItems(channelId)
}

func (s *ItemService) GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error) {
	ownedItems, err := s.db.GetOwnedItems(channelId, userId)
	if err != nil {
		return []models.Item{}, err
	}

	items := map[models.Item]bool{}
	for _, ownedItem := range ownedItems {
		items[ownedItem] = true
	}

	defaultItem, err := s.db.GetDefaultItem(channelId)
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
	return s.db.AddOwnedItem(userId, itemId, transactionId)
}
