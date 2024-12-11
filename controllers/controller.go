package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

type joinParams struct {
	UserID   models.TwitchID `json:"user_id"`
	Username string          `json:"username"`
}

type updateParams struct {
	ItemName string `json:"item_name"`
}

type Controller struct {
	announcer     services.Announcer
	authService   services.AuthServicer
	twitchRepo    repositories.TwitchRepository
	viewerService services.ViewerServicer
}

func NewOverlayController(
	announcer services.Announcer,
	authService services.AuthServicer,
	twitchRepo repositories.TwitchRepository,
	viewerService services.ViewerServicer,
) *Controller {
	return &Controller{
		announcer:     announcer,
		authService:   authService,
		twitchRepo:    twitchRepo,
		viewerService: viewerService,
	}
}

func (c *Controller) HandleListen(ctx *gin.Context) {
	channelID := models.TwitchID(ctx.Query("channelID"))
	overlayID, err := uuid.Parse(ctx.Query("overlayID"))
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.authService.VerifyOverlayID(channelID, overlayID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.twitchRepo.GetUsername(channelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	client := c.announcer.AddClient(channelName)
	defer func() {
		go func() {
			for range client.Stream {
			}
		}()
		c.announcer.RemoveClient(client)
	}()

	ctx.Stream(func(w io.Writer) bool {
		if event, ok := <-client.Stream; ok {
			ctx.SSEvent(event.Event, event.Message)
			return true
		}
		return false
	})
}

func (c *Controller) AddViewerToChannel(ctx *gin.Context) {
	channelName := ctx.Param("channelName")

	var joinParams joinParams
	if err := ctx.ShouldBindJSON(&joinParams); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	viewer, err := c.viewerService.GetViewer(joinParams.UserID, channelName, joinParams.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.announcer.AnnounceJoin(channelName, viewer)
}

func (c *Controller) RemoveViewerFromChannel(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	userID := models.TwitchID(ctx.Param("userID"))

	c.announcer.AnnouncePart(channelName, userID)
}

func (c *Controller) Action(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	action := ctx.Param("action")
	userID := models.TwitchID(ctx.Param("userID"))

	c.announcer.AnnounceAction(channelName, action, userID)
}

func (c *Controller) UpdateViewer(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	userID := models.TwitchID(ctx.Param("userID"))

	var updateParams updateParams
	if err := ctx.ShouldBindJSON(&updateParams); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.viewerService.UpdateViewer(userID, channelName, updateParams.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.announcer.AnnounceUpdate(channelName, item.Image, userID)
}

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
