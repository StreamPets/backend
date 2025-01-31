package controllers

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/models"
)

type clientAddRemover interface {
	AddClient(channelId models.TwitchId) announcers.Client
	RemoveClient(client announcers.Client)
}

type OverlayIdVerifier interface {
	VerifyOverlayId(channelId models.TwitchId, overlayId uuid.UUID) error
}

type OverlayController struct {
	announcer clientAddRemover
	Overlay   OverlayIdVerifier
}

func NewOverlayController(
	announcer clientAddRemover,
	overlay OverlayIdVerifier,
) *OverlayController {
	return &OverlayController{
		announcer: announcer,
		Overlay:   overlay,
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

	client := c.announcer.AddClient(channelId)
	defer func() {
		go func() {
			for range client.Stream {
			}
		}()
		c.announcer.RemoveClient(client)
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
