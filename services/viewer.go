package services

import "github.com/streampets/backend/models"

type Viewer struct {
	UserID   models.TwitchID
	Username string
	Image    string
}

type ItemRepo interface {
	GetSelectedItem(userID, channelID models.TwitchID) (models.Item, error)
}

type ViewerService struct {
	Items ItemRepo
}

func NewViewerService(
	items ItemRepo,
) *ViewerService {
	return &ViewerService{
		Items: items,
	}
}

func (s *ViewerService) GetViewer(userID, channelID models.TwitchID, username string) (Viewer, error) {
	item, err := s.Items.GetSelectedItem(userID, channelID)
	if err != nil {
		return Viewer{}, err
	}

	return Viewer{UserID: userID, Username: username, Image: item.Image}, nil
}
