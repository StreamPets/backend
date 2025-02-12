package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/auth"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/twitch"
)

func handleLogin(
	validateToken func(ctx context.Context, accessToken string) (twitch.Id, error),
	getOverlayId func(channelId twitch.Id) (uuid.UUID, error),
) gin.HandlerFunc {

	type userData struct {
		OverlayId uuid.UUID `json:"overlay_id"`
		ChannelId twitch.Id `json:"channel_id"`
	}

	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("Authorization")
		if authCookieErrorHandler(ctx, err) {
			return
		}

		userId, err := validateToken(ctx, token)
		if validateTokenErrorHandler(ctx, err) {
			return
		}

		overlayId, err := getOverlayId(userId)
		if getOverlayIdErrorHandler(ctx, err) {
			return
		}

		ctx.JSON(http.StatusOK, userData{
			OverlayId: overlayId,
			ChannelId: userId,
		})
	}
}

func handleListen(
	addClient func(channelId twitch.Id) announcers.Client,
	removeClient func(client announcers.Client),
	validateOverlayId func(channelId twitch.Id, overlayId uuid.UUID) error,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Query(ChannelId))
		overlayId, err := uuid.Parse(ctx.Query(OverlayId))
		if parseUuidErrorHandler(ctx, err) {
			return
		}

		err = validateOverlayId(channelId, overlayId)
		if validateOverlayIdErrorHandler(ctx, err) {
			return
		}

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
	verifyExtToken func(tokenString string) (*auth.ExtToken, error),
	getChannelsItems func(channelId twitch.Id) ([]models.Item, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := verifyExtToken(ctx.GetHeader(XExtensionJwt))
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		storeItems, err := getChannelsItems(token.ChannelId)
		if getChannelsItemsErrorHandler(ctx, err) {
			return
		}

		ctx.JSON(http.StatusOK, storeItems)
	}
}

func handleGetUserData(
	verifyExtToken func(tokenString string) (*auth.ExtToken, error),
	getSelectedItem func(userId, channelId twitch.Id) (models.Item, error),
	getOwnedItems func(channelId, userId twitch.Id) ([]models.Item, error),
) gin.HandlerFunc {

	type response struct {
		Selected models.Item   `json:"selected"`
		Owned    []models.Item `json:"owned"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		ownedItems, err := getOwnedItems(token.ChannelId, token.UserId)
		if getOwnedItemsErrorHandler(ctx, err) {
			return
		}

		selectedItem, err := getSelectedItem(token.UserId, token.ChannelId)
		if getSelectedItemErrorHandler(ctx, err) {
			return
		}

		ctx.JSON(http.StatusOK, response{
			Selected: selectedItem,
			Owned:    ownedItems,
		})
	}
}

func handleBuyStoreItem(
	verifyExtToken func(tokenString string) (*auth.ExtToken, error),
	verifyReceipt func(receiptString string) (*auth.Receipt, error),
	getItemById func(itemId uuid.UUID) (models.Item, error),
	addOwnedItem func(userId twitch.Id, itemId, transactionId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		Receipt string `json:"receipt"`
		ItemId  string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		request := new(request)
		err = ctx.ShouldBindJSON(request)
		if shouldBindJsonErrorHandler(ctx, err) {
			return
		}

		receipt, err := verifyReceipt(request.Receipt)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if parseUuidErrorHandler(ctx, err) {
			return
		}

		item, err := getItemById(itemId)
		if getItemByIdErrorHandler(ctx, err) {
			return
		}

		if item.Rarity != receipt.Data.Product.Rarity {
			slog.Error("receipt and item rarity do not match")
			ctx.JSON(http.StatusForbidden, nil)
			return
		}

		err = addOwnedItem(token.UserId, itemId, receipt.Data.TransactionId)
		if addOwnedItemErrorHandler(ctx, err) {
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleSetSelectedItem(
	announceUpdate func(channelId, userId twitch.Id, image string),
	verifyExtToken func(tokenString string) (*auth.ExtToken, error),
	getItemById func(itemId uuid.UUID) (models.Item, error),
	setSelectedItem func(userId, channelId twitch.Id, itemId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		ItemId string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		request := new(request)
		err = ctx.ShouldBindJSON(request)
		if shouldBindJsonErrorHandler(ctx, err) {
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if parseUuidErrorHandler(ctx, err) {
			return
		}

		item, err := getItemById(itemId)
		if getItemByIdErrorHandler(ctx, err) {
			return
		}

		err = setSelectedItem(token.UserId, token.ChannelId, itemId)
		if setSelectedItemErrorHandler(ctx, err) {
			return
		}

		announceUpdate(token.ChannelId, token.UserId, item.Image)
	}
}

func handleAddPetToChannel(
	announceJoin func(channelId twitch.Id, pet pets.Pet),
	getPet func(userId, channelId twitch.Id, username string) (pets.Pet, error),
) gin.HandlerFunc {

	type request struct {
		UserId   twitch.Id `json:"user_id"`
		Username string    `json:"username"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		err := ctx.ShouldBindJSON(request)
		if shouldBindJsonErrorHandler(ctx, err) {
			return
		}

		channelId := twitch.Id(ctx.Param(ChannelId))
		pet, err := getPet(request.UserId, channelId, request.Username)
		if getPetErrorHandler(ctx, err) {
			return
		}

		announceJoin(channelId, pet)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleRemoveUserFromChannel(
	announcePart func(channelId, userId twitch.Id),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Param(ChannelId))
		userId := twitch.Id(ctx.Param(UserId))

		announcePart(channelId, userId)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleAction(
	announceAction func(channelId, userId twitch.Id, action string),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Param(ChannelId))
		userId := twitch.Id(ctx.Param(UserId))
		action := ctx.Param(Action)

		announceAction(channelId, userId, action)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleUpdate(
	announceUpdate func(channelId, userId twitch.Id, image string),
	getItemByName func(channelId twitch.Id, itemName string) (models.Item, error),
	setSelectedItem func(userId, channelId twitch.Id, itemId uuid.UUID) error,
) gin.HandlerFunc {

	type request struct {
		ItemName string `json:"item_name"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		err := ctx.ShouldBindJSON(request)
		if shouldBindJsonErrorHandler(ctx, err) {
			return
		}

		channelId := twitch.Id(ctx.Param(ChannelId))
		userId := twitch.Id(ctx.Param(UserId))

		item, err := getItemByName(channelId, request.ItemName)
		if getItemByNameErrorHandler(ctx, err) {
			return
		}

		err = setSelectedItem(userId, channelId, item.ItemId)
		if setSelectedItemErrorHandler(ctx, err) {
			return
		}

		announceUpdate(channelId, userId, item.Image)
		ctx.JSON(http.StatusNoContent, nil)
	}
}
