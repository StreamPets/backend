package services

import "github.com/streampets/backend/models"

type Pet struct {
	UserId   models.TwitchId `json:"userId"`
	Username string          `json:"username"`
	Image    string          `json:"color"`
}

type ItemRepo interface {
	GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error)
}

type PetService struct {
	Items ItemRepo
}

func NewPetService(
	items ItemRepo,
) *PetService {
	return &PetService{
		Items: items,
	}
}

func (s *PetService) GetPet(userId, channelId models.TwitchId, username string) (Pet, error) {
	item, err := s.Items.GetSelectedItem(userId, channelId)
	if err != nil {
		return Pet{}, err
	}

	return Pet{UserId: userId, Username: username, Image: item.Image}, nil
}
