package pets

import (
	"github.com/streampets/backend/models"
)

type Pet struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Image    string `json:"color"`
}

type SelectedItemGetter interface {
	GetSelectedItem(userId, channelId string) (models.Item, error)
}

type PetService struct {
	items SelectedItemGetter
}

func New(items SelectedItemGetter) *PetService {
	return &PetService{items: items}
}

func (s *PetService) GetPet(userId, channelId string, username string) (Pet, error) {
	item, err := s.items.GetSelectedItem(userId, channelId)
	if err != nil {
		return Pet{}, err
	}

	return Pet{UserId: userId, Username: username, Image: item.Image}, nil
}
