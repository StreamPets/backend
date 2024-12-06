package services

import (
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
)

type Viewer struct {
	UserID   models.TwitchID
	Username string
	Image    string
}

type ViewerServicer interface {
	GetViewer(userID models.TwitchID, channelName, username string) (Viewer, error)
	UpdateViewer(userID models.TwitchID, channelName, itemName string) (models.Item, error)
}

type ViewerService struct {
	itemRepo   repositories.ItemRepository
	twitchRepo repositories.Twitcher
}

func NewViewerService(
	itemRepo repositories.ItemRepository,
	twitchRepo repositories.Twitcher,
) *ViewerService {
	return &ViewerService{
		itemRepo:   itemRepo,
		twitchRepo: twitchRepo,
	}
}

func (s *ViewerService) GetViewer(userID models.TwitchID, channelName, username string) (Viewer, error) {
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

func (s *ViewerService) UpdateViewer(userID models.TwitchID, channelName, itemName string) (models.Item, error) {
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
