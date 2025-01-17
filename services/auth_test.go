package services

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestVerifyOverlayId(t *testing.T) {
	t.Run("verify overlay id returns nil when ids match", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		overlayId := uuid.New()

		repoMock := mock.Mock[OverlayIdGetter]()
		mock.When(repoMock.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		authService := NewAuthService(repoMock, "")

		if err := authService.VerifyOverlayId(channelId, overlayId); err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayId(channelId)
	})

	t.Run("verify overlay id returns an error when ids do not match", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")

		repoMock := mock.Mock[OverlayIdGetter]()
		mock.When(repoMock.GetOverlayId(channelId)).ThenReturn(uuid.New(), nil)

		authService := NewAuthService(repoMock, "")

		if err := authService.VerifyOverlayId(channelId, uuid.New()); err == nil {
			t.Errorf("expected an error, but did not received one")
		} else if err != ErrIdMismatch {
			t.Errorf("expected %s got %s", err.Error(), ErrIdMismatch.Error())
		}

		mock.Verify(repoMock, mock.Once()).GetOverlayId(channelId)
	})
}

func TestVerifyExtToken(t *testing.T) {
	t.Run("valid token is verified correctly", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelId,
			"user_id":    userId,
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIdGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		got, err := authService.VerifyExtToken(tokenString)
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		if got.ChannelId != channelId {
			t.Errorf("expected %s got %s", channelId, got.ChannelId)
		}
		if got.UserId != userId {
			t.Errorf("expected %s got %s", userId, got.UserId)
		}
	})

	t.Run("invalid token is not verified", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelId,
			"user_id":    userId,
		})

		tokenString, err := token.SignedString([]byte("fake secret"))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIdGetter]()
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
		transactionId := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"data": map[string]interface{}{
				"transactionId": transactionId,
				"product": map[string]interface{}{
					"sku": "common",
				},
			},
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIdGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		got, err := authService.VerifyReceipt(tokenString)
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		if got.Data.TransactionId != transactionId {
			t.Errorf("expected %s got %s", transactionId, got.Data.TransactionId)
		}
	})

	t.Run("invalid token is not verified", func(t *testing.T) {
		mock.SetUp(t)

		clientSecret := "secret"
		transactionId := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"transaction_id": transactionId,
		})

		tokenString, err := token.SignedString([]byte("fake secret"))
		if err != nil {
			t.Errorf("did not expect an error but received %s", err.Error())
		}

		repoMock := mock.Mock[OverlayIdGetter]()
		authService := NewAuthService(repoMock, clientSecret)

		if _, err = authService.VerifyReceipt(tokenString); err == nil {
			t.Errorf("expected an error but did not received one")
		}
	})
}
