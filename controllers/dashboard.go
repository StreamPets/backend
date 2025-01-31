package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
)

type userData struct {
	OverlayId uuid.UUID       `json:"overlay_id"`
	ChannelId models.TwitchId `json:"channel_id"`
}

type DashboardController struct {
	GetOverlayId  func(channelId models.TwitchId) (overlayId uuid.UUID, err error)
	ValidateToken func(accessToken string) (response *twitch.ValidateTokenResponse, err error)
}

func NewDashboardController(
	GetOverlayId func(models.TwitchId) (uuid.UUID, error),
	ValidateToken func(string) (*twitch.ValidateTokenResponse, error),
) *DashboardController {
	return &DashboardController{
		GetOverlayId:  GetOverlayId,
		ValidateToken: ValidateToken,
	}
}

func (c *DashboardController) HandleLogin(ctx *gin.Context) {
	token, err := ctx.Cookie("Authorization")
	if err == http.ErrNoCookie {
		slog.Debug("no 'Authorization' cookie present")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	} else if err != nil {
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
	if err != nil {
		slog.Error("error when getting overlay url", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, userData{
		OverlayId: overlayId,
		ChannelId: response.UserId,
	})
}
