package controllers

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type ClientAddRemover interface {
	AddClient(channelName string) services.Client
	RemoveClient(client services.Client)
}

type OverlayIdVerifier interface {
	VerifyOverlayId(channelId models.TwitchId, overlayId uuid.UUID) error
}

type UsernameGetter interface {
	GetUsername(twitchId models.TwitchId) (string, error)
}

type OverlayController struct {
	Clients ClientAddRemover
	Overlay OverlayIdVerifier
	Users   UsernameGetter
}

func NewOverlayController(
	clients ClientAddRemover,
	overlay OverlayIdVerifier,
	users UsernameGetter,
) *OverlayController {
	return &OverlayController{
		Clients: clients,
		Overlay: overlay,
		Users:   users,
	}
}

func (c *OverlayController) HandleListen(ctx *gin.Context) {
	channelId := models.TwitchId(ctx.Query(ChannelId))
	overlayId, err := uuid.Parse(ctx.Query(OverlayId))
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.Overlay.VerifyOverlayId(channelId, overlayId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.Users.GetUsername(channelId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	client := c.Clients.AddClient(channelName)
	defer func() {
		go func() {
			for range client.Stream {
			}
		}()
		c.Clients.RemoveClient(client)
	}()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	ctx.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-client.Stream:
			if ok {
				ctx.SSEvent(event.Event, event.Message)
				return true
			}
			return false
		case <-ticker.C:
			ctx.SSEvent("heartbeat", "ping")
			return true
		}
	})
}
