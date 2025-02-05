package controllers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/twitch"
)

type OverlayIdGetter interface {
	GetOverlayId(channelId twitch.Id) (uuid.UUID, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, accessToken string) (twitch.Id, error)
}

func HandleLogin(tokens TokenValidator, overlays OverlayIdGetter) gin.HandlerFunc {
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
