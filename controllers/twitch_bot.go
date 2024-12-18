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

type ViewerGetter interface {
	GetViewer(userID, channelID models.TwitchID, username string) (services.Viewer, error)
}

type ItemGetterSetter interface {
	GetItemByName(channelID models.TwitchID, itemName string) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error
}

type UserIDGetter interface {
	GetUserID(username string) (models.TwitchID, error)
}

type TwitchBotController struct {
	Announcer Announcer
	Items     ItemGetterSetter
	Viewers   ViewerGetter
	Users     UserIDGetter
}

func NewTwitchBotController(
	announcer Announcer,
	items ItemGetterSetter,
	viewers ViewerGetter,
	users UserIDGetter,
) *TwitchBotController {
	return &TwitchBotController{
		Announcer: announcer,
		Items:     items,
		Viewers:   viewers,
		Users:     users,
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

	channelName := ctx.Param(ChannelName)
	channelID, err := c.Users.GetUserID(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	viewer, err := c.Viewers.GetViewer(params.UserID, channelID, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceJoin(channelName, viewer)
}

func (c *TwitchBotController) RemoveViewerFromChannel(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	userID := models.TwitchID(ctx.Param(UserID))

	c.Announcer.AnnouncePart(channelName, userID)
}

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	action := ctx.Param(Action)
	userID := models.TwitchID(ctx.Param(UserID))

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

	channelName := ctx.Param(ChannelName)
	userID := models.TwitchID(ctx.Param(UserID))

	channelID, err := c.Users.GetUserID(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.Items.GetItemByName(channelID, params.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Items.SetSelectedItem(userID, channelID, item.ItemID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, userID)
}
