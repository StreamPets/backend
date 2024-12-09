package services

import (
	"time"

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
	UpdateViewer(userID models.TwitchID, channelName, itemName string) (models.Item, error)
	GetTodaysItems(channelID models.TwitchID) ([]models.Item, error)
	GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error)
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

func (s *databaseService) UpdateViewer(userID models.TwitchID, channelName, itemName string) (models.Item, error) {
	channelID, err := s.twitchRepo.GetUserID(channelName)
	if err != nil {
		return models.Item{}, err
	}

	item, err := s.itemRepo.GetItemByName(channelID, itemName)
	if err != nil {
		return models.Item{}, err
	}

	if err := s.itemRepo.SetSelectedItem(channelID, userID, item.ItemID); err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (s *databaseService) GetTodaysItems(channelID models.TwitchID) ([]models.Item, error) {
	currentTime := time.Now()
	dayOfWeek := models.DayOfWeek(currentTime.Weekday().String())

	return s.itemRepo.GetScheduledItems(channelID, dayOfWeek)
}

func (s *databaseService) GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error) {
	return s.itemRepo.GetOwnedItems(channelID, userID)
}
