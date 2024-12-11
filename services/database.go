package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
)

type Viewer struct {
	UserID   models.TwitchID
	Username string
	Image    string
}

type DatabaseService interface {
	GetViewer(userID models.TwitchID, channelName, username string) (Viewer, error)

	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
	GetItemByID(itemID uuid.UUID) (models.Item, error)

	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error
	GetTodaysItems(channelID models.TwitchID) ([]models.Item, error)
	GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error)
	AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error
}

type databaseService struct {
	itemRepo   repositories.ItemRepository
	twitchRepo repositories.TwitchRepository
}

func NewDatabaseService(
	itemRepo repositories.ItemRepository,
	twitchRepo repositories.TwitchRepository,
) DatabaseService {
	return &databaseService{
		itemRepo:   itemRepo,
		twitchRepo: twitchRepo,
	}
}

func (s *databaseService) GetViewer(userID models.TwitchID, channelName, username string) (Viewer, error) {
	channelID, err := s.twitchRepo.GetUserID(channelName)
	if err != nil {
		return Viewer{}, nil
	}

	item, err := s.itemRepo.GetSelectedItem(userID, channelID)
	if err != nil {
		return Viewer{}, err
	}

	return Viewer{UserID: userID, Username: username, Image: item.Image}, nil
}

func (s *databaseService) GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error) {
	return s.itemRepo.GetItemByName(channelID, itemName)
}

func (s *databaseService) GetItemByID(itemID uuid.UUID) (models.Item, error) {
	return s.itemRepo.GetItemByID(itemID)
}

func (s *databaseService) SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error {
	return s.itemRepo.SetSelectedItem(channelID, userID, itemID)
}

func (s *databaseService) GetTodaysItems(channelID models.TwitchID) ([]models.Item, error) {
	currentTime := time.Now()
	dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())

	return s.itemRepo.GetScheduledItems(channelID, dayOfWeek)
}

func (s *databaseService) GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error) {
	return s.itemRepo.GetOwnedItems(channelID, userID)
}

func (s *databaseService) AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error {
	return s.itemRepo.AddOwnedItem(userID, itemID, transactionID)
}
