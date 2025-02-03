package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/twitch"
)

type userData struct {
	OverlayId uuid.UUID       `json:"overlay_id"`
	ChannelId models.TwitchId `json:"channel_id"`
}

type OverlayIdGetter interface {
	GetOverlayId(channelId models.TwitchId) (overlayId uuid.UUID, err error)
}

type TokenValidator interface {
	ValidateToken(accessToken string) (response *twitch.ValidateTokenResponse, err error)
}

type DashboardController struct {
	OverlayIdGetter
	TokenValidator
}

func NewDashboardController(
	overlayIdGetter OverlayIdGetter,
	tokenValidator TokenValidator,
) *DashboardController {
	return &DashboardController{
		OverlayIdGetter: overlayIdGetter,
		TokenValidator:  tokenValidator,
	}
}

func (c *DashboardController) HandleLogin(ctx *gin.Context) {
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

	response, err := c.ValidateToken(token)
	if err == twitch.ErrInvalidAccessToken {
		slog.Debug("invalid access token in header")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	} else if err != nil {
		slog.Error("error when validating access token", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	overlayId, err := c.GetOverlayId(response.UserId)
	if err == repositories.ErrNoOverlayId {
		slog.Error("no overlay id associated with channel id", "channel_id", response.UserId)
		ctx.JSON(http.StatusBadRequest, nil)
		return
	} else if err != nil {
		slog.Error("error when getting overlay url", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, userData{
		OverlayId: overlayId,
		ChannelId: response.UserId,
	})
}
