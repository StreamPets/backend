package routes

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/database"
	"github.com/streampets/backend/items"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/pets"
)

func handleLogin(
	getOverlayId func(channelId string) (overlayId uuid.UUID, err error),
) gin.HandlerFunc {

	type userData struct {
		OverlayId uuid.UUID `json:"overlay_id"`
		ChannelId string    `json:"channel_id"`
	}

	return func(ctx *gin.Context) {
		channelId := ctx.GetString(ChannelId)

		overlayId, err := getOverlayId(channelId)
		if errors.Is(err, database.ErrNoOverlayId) {
			slog.Error("no overlay id associated with channel id", "channel id", channelId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		} else if err != nil {
			slog.Error("error when getting overlay url", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, userData{
			OverlayId: overlayId,
			ChannelId: channelId,
		})
	}
}

func handleListen(
	addClient func(channelId string) announcers.Client,
	removeClient func(client announcers.Client),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := ctx.GetString(ChannelId)
		client := addClient(channelId)

		defer func() {
			go func() {
				for range client.Stream {
				}
			}()
			removeClient(client)
		}()

		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		ctx.Stream(func(w io.Writer) bool {
			select {
			case announcement, ok := <-client.Stream:
				if ok {
					ctx.SSEvent(announcement.Event, announcement.Message)
					return true
				}
				return false
			case <-ticker.C:
				ctx.SSEvent("heartbeat", "ping")
				return true
			}
		})
	}
}

func handleGetStoreData(
	getChannelsItems func(channelId string) ([]models.Item, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := ctx.GetString(ChannelId)

		storeItems, err := getChannelsItems(channelId)
		if err != nil {
			slog.Error("failed to retrieve channels items")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, storeItems)
	}
}

func handleGetUserData(
	getSelectedItem func(userId, channelId string) (models.Item, error),
	getOwnedItems func(channelId, userId string) ([]models.Item, error),
) gin.HandlerFunc {

	type response struct {
		Selected models.Item   `json:"selected"`
		Owned    []models.Item `json:"owned"`
	}

	return func(ctx *gin.Context) {
		channelId := ctx.GetString(ChannelId)
		userId := ctx.GetString(UserId)

		ownedItems, err := getOwnedItems(channelId, userId)
		if err != nil {
			slog.Error("failed to retrieve owned items", "channel id", channelId, "user id", userId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		selectedItem, err := getSelectedItem(userId, channelId)
		if err != nil {
			slog.Error("failed to retrieve selected item", "channel id", channelId, "user id", userId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, response{
			Selected: selectedItem,
			Owned:    ownedItems,
		})
	}
}

func handleBuyStoreItem(
	getItemById func(itemId uuid.UUID) (models.Item, error),
	addOwnedItem func(userId string, itemId, transactionId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		ItemId string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		userId := ctx.GetString(UserId)
		rarity := models.Rarity(ctx.GetString(Rarity))

		request := new(request)
		if err := ctx.ShouldBindJSON(request); err != nil {
			slog.Warn("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if err != nil {
			slog.Warn("could not parse item id to uuid", "item id", request.ItemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		transactionId, err := uuid.Parse(ctx.GetString(TransactionId))
		if err != nil {
			slog.Warn("could not parse transaction id to uuid", "transaction id", ctx.GetString(TransactionId))
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		item, err := getItemById(itemId)
		if errors.Is(err, database.ErrItemNotFoundById) {
			slog.Error("failed to retrieve item", "item id", itemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		} else if err != nil {
			slog.Error("error when retrieving item")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		if item.Rarity != rarity {
			slog.Error("rarities do not match", "item rarity", item.Rarity, "receipt rarity", rarity)
			ctx.JSON(http.StatusForbidden, nil)
			return
		}

		err = addOwnedItem(userId, itemId, transactionId)
		if errors.Is(err, database.ErrAddItemNotExist) {
			slog.Error("failed to add owned item", "user id", userId, "item id", itemId, "transaction id", transactionId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		} else if err != nil {
			slog.Error("failed to add owned item")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleSetSelectedItem(
	announceUpdate func(channelId, userId string, image string),
	getItemById func(itemId uuid.UUID) (models.Item, error),
	setSelectedItem func(userId, channelId string, itemId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		ItemId string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		channelId := ctx.GetString(ChannelId)
		userId := ctx.GetString(UserId)

		request := new(request)
		if err := ctx.ShouldBindJSON(request); err != nil {
			slog.Warn("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if err != nil {
			slog.Debug("item id is not uuid type", "item id", request.ItemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		item, err := getItemById(itemId)
		if errors.Is(err, database.ErrItemNotFoundById) {
			slog.Error("failed to retrieve item", "item id", itemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		} else if err != nil {
			slog.Error("error when retrieving item", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		err = setSelectedItem(userId, channelId, itemId)
		if errors.Is(err, items.ErrSelectUnownedItem) {
			slog.Error("user tried to select an item they did not own", "user id", userId, "channel id", channelId, "item id", itemId)
			ctx.JSON(http.StatusForbidden, nil)
			return
		} else if err != nil {
			slog.Error("failed to select item", "user id", userId, "channel id", channelId, "item id", itemId, "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		announceUpdate(channelId, userId, item.Image)
	}
}

func handleAddPetToChannel(
	announceJoin func(channelId string, pet pets.Pet),
	getPet func(userId, channelId string, username string) (pets.Pet, error),
) gin.HandlerFunc {

	type request struct {
		UserId   string `json:"user_id"`
		Username string `json:"username"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		err := ctx.ShouldBindJSON(request)
		if err != nil {
			slog.Warn("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		channelId := string(ctx.Param(ChannelId))
		pet, err := getPet(request.UserId, channelId, request.Username)
		if err != nil {
			slog.Error("failed to retrieve pet")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		announceJoin(channelId, pet)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleRemoveUserFromChannel(
	announcePart func(channelId, userId string),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := string(ctx.Param(ChannelId))
		userId := string(ctx.Param(UserId))

		announcePart(channelId, userId)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleAction(
	announceAction func(channelId, userId string, action string),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := string(ctx.Param(ChannelId))
		userId := string(ctx.Param(UserId))
		action := ctx.Param(Action)

		announceAction(channelId, userId, action)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleUpdate(
	announceUpdate func(channelId, userId string, image string),
	getItemByName func(channelId string, itemName string) (models.Item, error),
	setSelectedItem func(userId, channelId string, itemId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		ItemName string `json:"item_name"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		err := ctx.ShouldBindJSON(request)
		if err != nil {
			slog.Warn("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		channelId := string(ctx.Param(ChannelId))
		userId := string(ctx.Param(UserId))

		item, err := getItemByName(channelId, request.ItemName)
		if errors.Is(err, database.ErrItemNotFoundByName) {
			slog.Warn("item could not be found", "channel id", channelId, "item name", request.ItemName)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		} else if err != nil {
			slog.Error("error when retrieving item", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		err = setSelectedItem(userId, channelId, item.ItemId)
		if errors.Is(err, items.ErrSelectUnownedItem) {
			slog.Error("user tried to select an item they did not own")
			ctx.JSON(http.StatusForbidden, nil)
			return
		} else if err != nil {
			slog.Error("failed to select item")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		announceUpdate(channelId, userId, item.Image)
		ctx.JSON(http.StatusNoContent, nil)
	}
}
