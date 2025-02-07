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
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

const Action string = "action"
const XExtensionJwt string = "x-extension-jwt"

const ChannelId string = "channelId"
const OverlayId string = "overlayId"
const UserId string = "userId"

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

type itemByIdGetter interface {
	GetItemById(itemId uuid.UUID) (models.Item, error)
}

type ownedItemAdder interface {
	AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error
}

type foo interface {
	itemByIdGetter
	ownedItemAdder
}

type selectedItemSetter interface {
	SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error
}

type bar interface {
	itemByIdGetter
	selectedItemSetter
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

type joinAnnouncer interface {
	AnnounceJoin(channelId twitch.Id, pet services.Pet)
}

type partAnnouncer interface {
	AnnouncePart(channelId, userId twitch.Id)
}

type updateAnnouncer interface {
	AnnounceUpdate(channelId, userId twitch.Id, image string)
}

type petGetter interface {
	GetPet(userId, channelId twitch.Id, username string) (services.Pet, error)
}

func verifyExtTokenErrorHandler(ctx *gin.Context, err error) bool {
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

func validateTokenErrorHandler(ctx *gin.Context, err error) bool {
	if err == twitch.ErrInvalidUserToken {
		slog.Debug("invalid access token in header")
		ctx.JSON(http.StatusUnauthorized, nil)
		return true
	} else if err != nil {
		slog.Error("error when validating access token", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

func getOverlayIdErrorHandler(ctx *gin.Context, err error) bool {
	var e *repositories.ErrNoOverlayId
	if errors.As(err, &e) {
		slog.Error("no overlay id associated with channel id", "channel_id", e.ChannelId)
		ctx.JSON(http.StatusBadRequest, nil)
		return true
	} else if err != nil {
		slog.Error("error when getting overlay url", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

func authCookieErrorHandler(ctx *gin.Context, err error) bool {
	if err == http.ErrNoCookie {
		slog.Debug("no 'Authorization' cookie present")
		ctx.JSON(http.StatusUnauthorized, nil)
		return true
	} else if err != nil {
		// This should never happen since ctx.Cookie() only returns nil or http.ErrNoCookie.
		// If this does occur, it might indicate a bug.
		slog.Error("error when retrieving 'Authorization' cookie", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusForbidden [403] if err is an ErrSelectUnownedItem.
//
// Returns InternalServerError [500] otherwise.
func setSelectedItemErrorHandler(ctx *gin.Context, err error) bool {
	e := new(services.ErrSelectUnownedItem)
	if errors.As(err, e) {
		slog.Error("user tried to select an item they did not own", "user id", e.UserId, "channel id", e.ChannelId, "item id", e.ItemId)
		ctx.JSON(http.StatusForbidden, nil)
		return true
	} else if err != nil {
		slog.Error("failed to select item")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusBadRequest [400] if err is not nil.
func shouldBindJsonErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Warn("failed to bind json")
		ctx.JSON(http.StatusBadRequest, nil)
		return true
	}
	return false
}

// Returns StatusBadRequest [400] if err is not nil.
func getItemByNameErrorHandler(ctx *gin.Context, err error) bool {
	e := new(services.ErrItemNotFound)
	if errors.As(err, e) {
		slog.Warn("item could not be found", "item name", e.ItemName)
		ctx.JSON(http.StatusBadRequest, nil)
		return true
	} else if err != nil {
		slog.Error("error when retrieving item", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}
