package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type Announcer interface {
	AnnounceJoin(channelName string, viewer services.Viewer)
	AnnouncePart(channelName string, userID models.TwitchID)
	AnnounceAction(channelName, action string, userID models.TwitchID)
	AnnounceUpdate(channelName, image string, userID models.TwitchID)
}

// TODO: Should these be the same service?
type DBService interface {
	GetViewer(userID, channelID models.TwitchID, username string) (services.Viewer, error)
	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error
}

// TODO: Rename to ~ TwitchApi
type UserIDGetter interface {
	GetUserID(username string) (models.TwitchID, error)
}

type TwitchBotController struct {
	Announcer Announcer
	Database  DBService
	// Rename:
	UserIDs UserIDGetter
}

func NewTwitchBotController(
	announcer Announcer,
	database DBService,
	users UserIDGetter,
) *TwitchBotController {
	return &TwitchBotController{
		Announcer: announcer,
		Database:  database,
		UserIDs:   users,
	}
}

func (c *TwitchBotController) AddViewerToChannel(ctx *gin.Context) {
	type Params struct {
		UserID   models.TwitchID `json:"user_id"`
		Username string          `json:"username"`
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName := ctx.Param("channelName")
	channelID, err := c.UserIDs.GetUserID(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	viewer, err := c.Database.GetViewer(params.UserID, channelID, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceJoin(channelName, viewer)
}

func (c *TwitchBotController) RemoveViewerFromChannel(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	userID := models.TwitchID(ctx.Param("userID"))

	c.Announcer.AnnouncePart(channelName, userID)
}

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	action := ctx.Param("action")
	userID := models.TwitchID(ctx.Param("userID"))

	c.Announcer.AnnounceAction(channelName, action, userID)
}

func (c *TwitchBotController) UpdateViewer(ctx *gin.Context) {
	type Params struct {
		ItemName string `json:"item_name"`
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName := ctx.Param("channelName")
	userID := models.TwitchID(ctx.Param("userID"))

	channelID, err := c.UserIDs.GetUserID(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.Database.GetItemByName(channelID, params.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Database.SetSelectedItem(userID, channelID, item.ItemID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, userID)
}
