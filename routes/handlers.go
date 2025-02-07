package routes

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
)

func handleLogin(
	tokens tokenValidator,
	overlays overlayIdGetter,
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

		userId, err := tokens.ValidateToken(ctx, token)
		if validateTokenErrorHandler(ctx, err) {
			return
		}

		overlayId, err := overlays.GetOverlayId(userId)
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
	announcer clientAddRemover,
	overlay overlayIdValidator,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Query(ChannelId))
		overlayId, err := uuid.Parse(ctx.Query(OverlayId))
		if err != nil {
			slog.Debug("query param overlay id is not uuid type")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		if err := overlay.ValidateOverlayId(channelId, overlayId); err != nil {
			slog.Warn("unrecognised overlay id", "overlay id", overlayId)
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		client := announcer.AddClient(channelId)
		defer func() {
			go func() {
				for range client.Stream {
				}
			}()
			announcer.RemoveClient(client)
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
	verifier extTokenVerifier,
	store channelItemGetter,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifier.VerifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		storeItems, err := store.GetChannelsItems(token.ChannelId)
		if err != nil {
			slog.Error("failed to retrieve channels items", "channel id", token.ChannelId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, storeItems)
	}
}

func handleGetUserData(
	verifier extTokenVerifier,
	store userDataGetter,
) gin.HandlerFunc {

	type response struct {
		Selected models.Item   `json:"selected"`
		Owned    []models.Item `json:"owned"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifier.VerifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		ownedItems, err := store.GetOwnedItems(token.ChannelId, token.UserId)
		if err != nil {
			slog.Error("failed to retrieve owned items")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		selectedItem, err := store.GetSelectedItem(token.UserId, token.ChannelId)
		if err != nil {
			slog.Error("failed to retrieve selected item")
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
	verifier tokenVerifier,
	store foo,
) gin.HandlerFunc {

	type request struct {
		Receipt string `json:"receipt"`
		ItemId  string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifier.VerifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		request := new(request)
		if err = ctx.ShouldBindJSON(request); err != nil {
			slog.Error("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		receipt, err := verifier.VerifyReceipt(request.Receipt)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if err != nil {
			slog.Error("failed to parse item id", "item id", request.ItemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		item, err := store.GetItemById(itemId)
		if err != nil {
			slog.Error("failed to retrieve item", "item id", itemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		if item.Rarity != receipt.Data.Product.Rarity {
			slog.Error("receipt and item rarity do not match")
			ctx.JSON(http.StatusForbidden, nil)
			return
		}

		if err := store.AddOwnedItem(token.UserId, itemId, receipt.Data.TransactionId); err != nil {
			slog.Error("failed to add owned item", "user id", token.UserId, "item id", itemId, "channel id", token.ChannelId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleSetSelectedItem(
	announcer updateAnnouncer,
	verifier extTokenVerifier,
	store bar,
) gin.HandlerFunc {

	type request struct {
		ItemId string `json:"item_id"`
	}

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token, err := verifier.VerifyExtToken(tokenString)
		if verifyExtTokenErrorHandler(ctx, err) {
			return
		}

		request := new(request)
		if err = ctx.ShouldBindJSON(request); err != nil {
			slog.Error("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		itemId, err := uuid.Parse(request.ItemId)
		if err != nil {
			slog.Error("failed to parse item id", "item id", request.ItemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		item, err := store.GetItemById(itemId)
		if err != nil {
			slog.Error("failed to retrieve item", "item id", itemId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		err = store.SetSelectedItem(token.UserId, token.ChannelId, itemId)
		if setSelectedItemErrorHandler(ctx, err) {
			return
		}

		announcer.AnnounceUpdate(token.ChannelId, token.UserId, item.Image)
	}
}

func handleAddPetToChannel(
	announcer joinAnnouncer,
	pets petGetter,
) gin.HandlerFunc {

	type request struct {
		UserId   twitch.Id `json:"user_id"`
		Username string    `json:"username"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		err := ctx.ShouldBindJSON(request)
		if err != nil {
			slog.Error("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		channelId := twitch.Id(ctx.Param(ChannelId))
		pet, err := pets.GetPet(request.UserId, channelId, request.Username)
		if err != nil {
			slog.Error("failed to retrieve pet", "user id", request.UserId, "channel id", channelId)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		announcer.AnnounceJoin(channelId, pet)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleRemoveUserFromChannel(
	announcer partAnnouncer,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Param(ChannelId))
		userId := twitch.Id(ctx.Param(UserId))

		announcer.AnnouncePart(channelId, userId)
		ctx.JSON(http.StatusNoContent, nil)
	}
}

func handleAction(
	announcer actionAnnouncer,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelId := twitch.Id(ctx.Param(ChannelId))
		userId := twitch.Id(ctx.Param(UserId))
		action := ctx.Param(Action)

		announcer.AnnounceAction(channelId, userId, action)
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
