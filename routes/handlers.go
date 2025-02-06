package routes

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
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
		if err == http.ErrNoCookie {
			slog.Debug("no 'Authorization' cookie present")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			// This should never happen since ctx.Cookie() only returns nil or http.ErrNoCookie.
			// If this does occur, it might indicate a bug.
			slog.Error("error when retrieving 'Authorization' cookie", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		userId, err := tokens.ValidateToken(ctx, token)
		if err == twitch.ErrInvalidUserToken {
			slog.Debug("invalid access token in header")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			slog.Error("error when validating access token", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		overlayId, err := overlays.GetOverlayId(userId)
		var e *repositories.ErrNoOverlayId
		if errors.As(err, &e) {
			slog.Error("no overlay id associated with channel id", "channel_id", e.ChannelId)
			ctx.JSON(http.StatusBadRequest, nil)
			return
		} else if err != nil {
			slog.Error("error when getting overlay url", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
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
		if err == services.ErrInvalidToken {
			slog.Warn("invalid extension token", "token", tokenString)
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			slog.Error("failed to validate token", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
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
