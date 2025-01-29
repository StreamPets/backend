package controllers

import (
	"io"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type ClientAddRemover interface {
	AddClient(channelId models.TwitchId) services.Client
	RemoveClient(client services.Client)
}

type OverlayIdVerifier interface {
	VerifyOverlayId(channelId models.TwitchId, overlayId uuid.UUID) error
}

type OverlayController struct {
	Clients ClientAddRemover
	Overlay OverlayIdVerifier
}

func NewOverlayController(
	clients ClientAddRemover,
	overlay OverlayIdVerifier,
) *OverlayController {
	return &OverlayController{
		Clients: clients,
		Overlay: overlay,
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

	client := c.Clients.AddClient(channelId)
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
			slog.Info("received event %s: %s", event.Event, event.Message)
			if ok {
				ctx.SSEvent(event.Event, event.Message)
				return true
			}
			slog.Info("returning false...")
			return false
		case <-ticker.C:
			slog.Info("sending heartbeat")
			ctx.SSEvent("heartbeat", "ping")
			return true
		}
	})
}
