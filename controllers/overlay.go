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

type OverlayIDVerifier interface {
	VerifyOverlayID(channelID models.TwitchID, overlayID uuid.UUID) error
}

type UsernameGetter interface {
	GetUsername(userID models.TwitchID) (string, error)
}

type ViewersGetter interface {
	GetViewers(channelID models.TwitchID) []services.Viewer
}

type OverlayController struct {
	Clients ClientAddRemover
	Overlay OverlayIDVerifier
	Users   UsernameGetter
	Cache   ViewersGetter
}

func NewOverlayController(
	clients ClientAddRemover,
	overlay OverlayIDVerifier,
	users UsernameGetter,
	cache ViewersGetter,
) *OverlayController {
	return &OverlayController{
		Clients: clients,
		Overlay: overlay,
		Users:   users,
		Cache:   cache,
	}
}

func (c *OverlayController) HandleListen(ctx *gin.Context) {
	channelID := models.TwitchID(ctx.Query(ChannelID))
	overlayID, err := uuid.Parse(ctx.Query(OverlayID))
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.Overlay.VerifyOverlayID(channelID, overlayID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.Users.GetUsername(channelID)
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

	for _, viewer := range c.Cache.GetViewers(channelID) {
		ctx.SSEvent("JOIN", viewer)
	}

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
