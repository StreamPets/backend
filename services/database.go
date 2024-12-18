package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

type Viewer struct {
	UserID   models.TwitchID
	Username string
	Image    string
}

type ItemRepository interface {
	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
	GetItemByID(itemID uuid.UUID) (models.Item, error)

	GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error

	GetScheduledItems(channelID models.TwitchID, dayOfWeek models.DayOfWeek) ([]models.Item, error)

	GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error)
	AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error
	CheckOwnedItem(userID models.TwitchID, itemID uuid.UUID) error
}

type DatabaseService struct {
	itemRepo ItemRepository
}

func NewDatabaseService(
	itemRepo ItemRepository,
) *DatabaseService {
	return &DatabaseService{
		itemRepo: itemRepo,
	}
}

func (s *DatabaseService) GetViewer(userID, channelID models.TwitchID, username string) (Viewer, error) {
	// Check if userID exists in users table
	// Create user if not exists

	item, err := s.itemRepo.GetSelectedItem(userID, channelID)
	if err != nil {
		return Viewer{}, err
	}

	// Check if user has selected an item
	// Retrieve channel's default item if not

	return Viewer{UserID: userID, Username: username, Image: item.Image}, nil
}

func (s *DatabaseService) GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error) {
	return s.itemRepo.GetItemByName(channelID, itemName)
}

func (s *DatabaseService) GetItemByID(itemID uuid.UUID) (models.Item, error) {
	return s.itemRepo.GetItemByID(itemID)
}

func (s *DatabaseService) SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error {
	if err := s.itemRepo.CheckOwnedItem(userID, itemID); err != nil {
		return err
	}

	return s.itemRepo.SetSelectedItem(channelID, userID, itemID)
}

func (s *DatabaseService) GetTodaysItems(channelID models.TwitchID) ([]models.Item, error) {
	currentTime := time.Now()
	dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())

	return s.itemRepo.GetScheduledItems(channelID, dayOfWeek)
}

func (s *DatabaseService) GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error) {
	return s.itemRepo.GetOwnedItems(channelID, userID)
}

func (s *DatabaseService) AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(userID, itemID, transactionID)
}
