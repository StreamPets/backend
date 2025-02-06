package routes

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

const ChannelId string = "channelId"
const OverlayId string = "overlayId"
const XExtensionJwt string = "x-extension-jwt"

type overlayIdGetter interface {
	GetOverlayId(channelId twitch.Id) (uuid.UUID, error)
}

type overlayIdValidator interface {
	ValidateOverlayId(channelId twitch.Id, overlayId uuid.UUID) error
}

type tokenValidator interface {
	ValidateToken(ctx context.Context, accessToken string) (twitch.Id, error)
}

type clientAdder interface {
	AddClient(channelId twitch.Id) announcers.Client
}

type clientRemover interface {
	RemoveClient(client announcers.Client)
}

type clientAddRemover interface {
	clientAdder
	clientRemover
}

type extTokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
}

type receiptVerifier interface {
	VerifyReceipt(receiptString string) (*services.Receipt, error)
}

type tokenVerifier interface {
	extTokenVerifier
	receiptVerifier
}

type channelItemGetter interface {
	GetChannelsItems(channelId twitch.Id) ([]models.Item, error)
}

type itemIdGetter interface {
	GetItemById(itemId uuid.UUID) (models.Item, error)
}

type ownedItemAdder interface {
	AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error
}

type foo interface {
	itemIdGetter
	ownedItemAdder
}

type selectedItemGetter interface {
	GetSelectedItem(userId, channelId twitch.Id) (models.Item, error)
}

type ownedItemsGetter interface {
	GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error)
}

type userDataGetter interface {
	selectedItemGetter
	ownedItemsGetter
}

func verifierTokenErrorHandler(ctx *gin.Context, err error) bool {
	var e *services.ErrInvalidToken
	if errors.As(err, &e) {
		slog.Warn("invalid token", "token", e.TokenString)
		ctx.JSON(http.StatusUnauthorized, nil)
		return true
	} else if err != nil {
		slog.Error("failed to validate token", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}
