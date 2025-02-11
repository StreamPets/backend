package services

import (
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
)

type Pet struct {
	UserId   twitch.Id `json:"userId"`
	Username string    `json:"username"`
	Image    string    `json:"color"`
}

type SelectedItemGetter interface {
	GetSelectedItem(userId, channelId twitch.Id) (models.Item, error)
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

func (s *PetService) GetPet(userId, channelId twitch.Id, username string) (Pet, error) {
	item, err := s.items.GetSelectedItem(userId, channelId)
	if err != nil {
		return Pet{}, err
	}

	return Pet{UserId: userId, Username: username, Image: item.Image}, nil
}
