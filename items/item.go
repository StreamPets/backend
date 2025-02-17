package items

import (
	"github.com/google/uuid"
	streampets "github.com/streampets/backend"
)

type database interface {
	Item(itemId uuid.UUID) (streampets.Item, error)
	ItemByName(channelId string, itemName string) (streampets.Item, error)

	SelectedItem(userId, channelId string) (streampets.Item, error)
	SetSelectedItem(userId, channelId string, itemId uuid.UUID) error
	DeleteSelectedItem(userId, channelId string) error

	ItemsByChannelId(channelId string) ([]streampets.Item, error)
	ItemsByUserId(channelId, userId string) ([]streampets.Item, error)

	DefaultItem(channelId string) (streampets.Item, error)

	CreateOwnedItem(userId string, itemId, transactionId uuid.UUID) error
	ItemOwned(userId string, itemId uuid.UUID) (bool, error)
}

type ItemService struct {
	db database
}

func New(db database) *ItemService {
	return &ItemService{db: db}
}

func (s *ItemService) GetItemByName(channelId string, itemName string) (streampets.Item, error) {
	return s.db.ItemByName(channelId, itemName)
}

func (s *ItemService) GetItemById(itemId uuid.UUID) (streampets.Item, error) {
	return s.db.Item(itemId)
}

func (s *ItemService) GetSelectedItem(userId, channelId string) (streampets.Item, error) {
	item, err := s.db.SelectedItem(userId, channelId)
	if err != nil {
		return streampets.Item{}, err
	}

	return item, nil
}

func (s *ItemService) SetSelectedItem(userId, channelId string, itemId uuid.UUID) error {
	if owned, err := s.db.ItemOwned(userId, itemId); err != nil {
		return err
	} else if owned {
		return s.db.SetSelectedItem(channelId, userId, itemId)
	}

	if defaultItem, err := s.db.DefaultItem(channelId); err != nil {
		return err
	} else if defaultItem.ItemId != itemId {
		return ErrSelectUnownedItem
	}

	return s.db.DeleteSelectedItem(userId, channelId)
}

func (s *ItemService) GetChannelsItems(channelId string) ([]streampets.Item, error) {
	return s.db.ItemsByChannelId(channelId)
}

func (s *ItemService) GetOwnedItems(channelId, userId string) ([]streampets.Item, error) {
	ownedItems, err := s.db.ItemsByUserId(channelId, userId)
	if err != nil {
		return []streampets.Item{}, err
	}

	items := map[streampets.Item]bool{}
	for _, ownedItem := range ownedItems {
		items[ownedItem] = true
	}

	defaultItem, err := s.db.DefaultItem(channelId)
	if err != nil {
		return []streampets.Item{}, err
	}
	items[defaultItem] = true

	result := []streampets.Item{}
	for item := range items {
		result = append(result, item)
	}

	return result, nil
}

func (s *ItemService) AddOwnedItem(userId string, itemId, transactionId uuid.UUID) error {
	return s.db.CreateOwnedItem(userId, itemId, transactionId)
}
