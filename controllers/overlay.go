package controllers

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type ClientAddRemover interface {
	AddClient(channelName string) services.Client
	RemoveClient(client services.Client)
}

type OverlayIDVerifier interface {
	VerifyOverlayID(channelID models.TwitchID, overlayID uuid.UUID) error
}

type UsernameGetter interface {
	GetUsername(userID models.TwitchID) (string, error)
}

type OverlayController struct {
	clients   ClientAddRemover
	overlays  OverlayIDVerifier
	usernames UsernameGetter
}

func NewOverlayController(
	clients ClientAddRemover,
	overlays OverlayIDVerifier,
	usernames UsernameGetter,
) *OverlayController {
	return &OverlayController{
		clients:   clients,
		overlays:  overlays,
		usernames: usernames,
	}
}

func (c *OverlayController) HandleListen(ctx *gin.Context) {
	channelID := models.TwitchID(ctx.Query("channelID"))
	overlayID, err := uuid.Parse(ctx.Query("overlayID"))
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.overlays.VerifyOverlayID(channelID, overlayID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.usernames.GetUsername(channelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	client := c.clients.AddClient(channelName)
	defer func() {
		go func() {
			for range client.Stream {
			}
		}()
		c.clients.RemoveClient(client)
	}()

	ctx.Stream(func(w io.Writer) bool {
		if event, ok := <-client.Stream; ok {
			ctx.SSEvent(event.Event, event.Message)
			return true
		}
		return false
	})
}
