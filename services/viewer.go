package services

import "github.com/streampets/backend/models"

type Pet struct {
	ViewerId models.UserId `json:"viewerId"`
	Username string        `json:"username"`
	Image    string        `json:"color"`
}

type ItemRepo interface {
	GetSelectedItem(viewerId, channelId models.UserId) (models.Item, error)
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

func (s *PetService) GetPet(viewerId, channelId models.UserId, username string) (Pet, error) {
	item, err := s.Items.GetSelectedItem(viewerId, channelId)
	if err != nil {
		return Pet{}, err
	}

	return Pet{ViewerId: viewerId, Username: username, Image: item.Image}, nil
}
