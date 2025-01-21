package services

import "github.com/streampets/backend/models"

type Pet struct {
	UserId   models.TwitchId `json:"userId"`
	Username string          `json:"username"`
	Image    string          `json:"color"`
}

type SelectedItemGetter interface {
	GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error)
}

type PetService struct {
	items SelectedItemGetter
}

func NewPetService(
	items SelectedItemGetter,
) *PetService {
	return &PetService{
		items: items,
	}
}

func (s *PetService) GetPet(userId, channelId models.TwitchId, username string) (Pet, error) {
	item, err := s.items.GetSelectedItem(userId, channelId)
	if err != nil {
		return Pet{}, err
	}

	return Pet{UserId: userId, Username: username, Image: item.Image}, nil
}
