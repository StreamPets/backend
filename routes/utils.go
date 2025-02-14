package routes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/database"
	"github.com/streampets/backend/items"
)

const Action string = "action"
const XExtensionJwt string = "x-extension-jwt"

const ChannelId string = "channelId"
const OverlayId string = "overlayId"
const UserId string = "userId"
const ItemId string = "itemId"
const TransactionId string = "transactionId"
const Rarity string = "rarity"

// Returns StatusForbidden [403] if err is an ErrSelectUnownedItem.
//
// Returns InternalServerError [500] otherwise.
func setSelectedItemErrorHandler(ctx *gin.Context, err error) bool {
	e := new(items.ErrSelectUnownedItem)
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
	e := new(database.ErrItemNotFoundByName)
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

// Returns StatusInternalServerError [500] if err is not nil.
func getPetErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Error("failed to retrieve pet")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusBadRequest [400] if err is not nil.
func parseUuidErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Debug("param is not uuid type")
		ctx.JSON(http.StatusBadRequest, nil)
		return true
	}
	return false
}

// Returns StatusUnauthorized [401] if OverlayId and ChannelId do not match.
func validateOverlayIdErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Warn("unrecognised overlay id", "overlay id", ctx.Query(OverlayId), "channel id", ctx.Query(ChannelId))
		ctx.JSON(http.StatusUnauthorized, nil)
		return true
	}
	return false
}

// Returns StatusInternalServerError [500] if err is not nil.
func getOwnedItemsErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Error("failed to retrieve owned items")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusInternalServerError [500] if err is not nil.
func getSelectedItemErrorHandler(ctx *gin.Context, err error) bool {
	if err != nil {
		slog.Error("failed to retrieve selected item")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusBadRequest [400] if err is an ErrItemNotFoundById.
//
// Returns StatusInternalServerError [500] is err is not nil.
func getItemByIdErrorHandler(ctx *gin.Context, err error) bool {
	e := new(database.ErrItemNotFoundById)
	if errors.As(err, e) {
		slog.Error("failed to retrieve item", "item id", e.ItemId)
		ctx.JSON(http.StatusBadRequest, nil)
		return true
	} else if err != nil {
		slog.Error("error when retrieving item")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}

// Returns StatusInternalServerError [500] if err is an ErrAddItemNotExist.
//
// Returns StatusInternalServerError [500] if err is not nil.
func addOwnedItemErrorHandler(ctx *gin.Context, err error) bool {
	e := new(database.ErrAddItemNotExist)
	if errors.As(err, e) {
		slog.Error("failed to add owned item", "user id", e.UserId, "item id", e.ItemId, "transaction id", e.TransactionId)
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	} else if err != nil {
		slog.Error("failed to add owned item")
		ctx.JSON(http.StatusInternalServerError, nil)
		return true
	}
	return false
}
