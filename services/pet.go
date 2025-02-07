package services

import (
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
)

type Pet struct {
	UserId   twitch.Id `json:"userId"`
	Username string    `json:"username"`
	Image    string    `json:"color"`
}

// TODO: More informative name
type ErrCreatePet struct {
	UserId    twitch.Id
	Username  string
	ChannelId twitch.Id
}

// TODO: Should this return a pointer?
func NewErrCreatePet(
	userId twitch.Id,
	username string,
	channelId twitch.Id,
) ErrCreatePet {
	return ErrCreatePet{
		UserId:    userId,
		Username:  username,
		ChannelId: channelId,
	}
}

func (e ErrCreatePet) Error() string {
	return "failed to create pet"
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
	if err == gorm.ErrRecordNotFound {
		return Pet{}, NewErrCreatePet(userId, username, channelId)
	} else if err != nil {
		return Pet{}, err
	}

	return Pet{UserId: userId, Username: username, Image: item.Image}, nil
}
