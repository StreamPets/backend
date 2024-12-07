package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

var ErrIdMismatch = errors.New("channelID and overlayID do not match")

type ChannelRepo interface {
	GetOverlayID(channelID models.TwitchID) (uuid.UUID, error)
}

type AuthServicer interface {
	VerifyOverlayID(models.TwitchID, uuid.UUID) error
}

type AuthService struct {
	channelRepo ChannelRepo
}

func NewAuthService(channelRepo ChannelRepo) *AuthService {
	return &AuthService{channelRepo: channelRepo}
}

func (s *AuthService) VerifyOverlayID(channelID models.TwitchID, overlayID uuid.UUID) error {
	expectedID, err := s.channelRepo.GetOverlayID(channelID)
	if err != nil {
		return err
	}

	if overlayID != expectedID {
		return ErrIdMismatch
	}

	return nil
}
