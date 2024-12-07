package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

func TestVerifyOverlayID(t *testing.T) {
	t.Run("verify overlay id returns nil when ids match", func(t *testing.T) {
		channelID := models.TwitchID("channel id")
		overlayID := uuid.New()

		repoMock := mock.Mock[services.ChannelRepo]()
		mock.When(repoMock.GetOverlayID(channelID)).ThenReturn(overlayID, nil)

		authService := services.NewAuthService(repoMock)

		if err := authService.VerifyOverlayID(channelID, overlayID); err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayID(channelID)
	})

	t.Run("verify overlay id returns an error when ids do not match", func(t *testing.T) {
		channelID := models.TwitchID("channel id")

		repoMock := mock.Mock[services.ChannelRepo]()
		mock.When(repoMock.GetOverlayID(channelID)).ThenReturn(uuid.New(), nil)

		authService := services.NewAuthService(repoMock)

		if err := authService.VerifyOverlayID(channelID, uuid.New()); err == nil {
			t.Errorf("expected an error, but did not received one")
		} else if err != services.ErrIdMismatch {
			t.Errorf("expected %s got %s", err.Error(), services.ErrIdMismatch.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayID(channelID)
	})
}
