package services

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestVerifyOverlayID(t *testing.T) {
	t.Run("verify overlay id returns nil when ids match", func(t *testing.T) {
		mock.SetUp(t)

		channelID := models.TwitchID("channel id")
		overlayID := uuid.New()

		repoMock := mock.Mock[OverlayIDGetter]()
		mock.When(repoMock.GetOverlayID(channelID)).ThenReturn(overlayID, nil)

		authService := NewAuthService(repoMock, "")

		if err := authService.VerifyOverlayID(channelID, overlayID); err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayID(channelID)
	})

	t.Run("verify overlay id returns an error when ids do not match", func(t *testing.T) {
		mock.SetUp(t)

		channelID := models.TwitchID("channel id")

		repoMock := mock.Mock[OverlayIDGetter]()
		mock.When(repoMock.GetOverlayID(channelID)).ThenReturn(uuid.New(), nil)

		authService := NewAuthService(repoMock, "")

		if err := authService.VerifyOverlayID(channelID, uuid.New()); err == nil {
			t.Errorf("expected an error, but did not received one")
		} else if err != ErrIdMismatch {
			t.Errorf("expected %s got %s", err.Error(), ErrIdMismatch.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayID(channelID)
	})
}

func TestVerifyExtToken(t *testing.T) {
	t.Run("valid token is verified correctly", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		channelID := models.TwitchID("channel id")
		userID := models.TwitchID("user id")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelID,
			"user_id":    userID,
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIDGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		got, err := authService.VerifyExtToken(tokenString)
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		if got.ChannelID != channelID {
			t.Errorf("expected %s got %s", channelID, got.ChannelID)
		}
		if got.UserID != userID {
			t.Errorf("expected %s got %s", userID, got.UserID)
		}
	})

	t.Run("invalid token is not verified", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		channelID := models.TwitchID("channel id")
		userID := models.TwitchID("user id")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelID,
			"user_id":    userID,
		})

		tokenString, err := token.SignedString([]byte("fake secret"))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIDGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		if _, err = authService.VerifyExtToken(tokenString); err == nil {
			t.Errorf("expected an error but did not received one")
		}
	})
}

func TestVerifyReceipt(t *testing.T) {
	t.Run("valid token is verified correctly", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		transactionID := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"transaction_id": transactionID,
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIDGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		got, err := authService.VerifyReceipt(tokenString)
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		if got.TransactionID != transactionID {
			t.Errorf("expected %s got %s", transactionID, got.TransactionID)
		}
	})

	t.Run("invalid token is not verified", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		transactionID := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"transaction_id": transactionID,
		})

		tokenString, err := token.SignedString([]byte("fake secret"))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIDGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		if _, err = authService.VerifyReceipt(tokenString); err == nil {
			t.Errorf("expected an error but did not received one")
		}
	})
}
