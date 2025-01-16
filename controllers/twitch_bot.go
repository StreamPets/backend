package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type Announcer interface {
	AnnounceJoin(channelName string, viewer services.Pet)
	AnnouncePart(channelName string, viewerId models.TwitchId)
	AnnounceAction(channelName, action string, viewerId models.TwitchId)
	AnnounceUpdate(channelName, image string, viewerId models.TwitchId)
}

type PetGetter interface {
	GetPet(viewerId, channelId models.TwitchId, username string) (services.Pet, error)
}

type ItemGetSetter interface {
	GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error)
	SetSelectedItem(viewerId, channelId models.TwitchId, itemId uuid.UUID) error
}

type UserIdGetter interface {
	GetUserId(username string) (models.TwitchId, error)
}

type TwitchBotController struct {
	Announcer Announcer
	Items     ItemGetSetter
	Pets      PetGetter
	Users     UserIdGetter
}

func NewTwitchBotController(
	announcer Announcer,
	items ItemGetSetter,
	pets PetGetter,
	users UserIdGetter,
) *TwitchBotController {
	return &TwitchBotController{
		Announcer: announcer,
		Items:     items,
		Pets:      pets,
		Users:     users,
	}
}

func (c *TwitchBotController) AddViewerToChannel(ctx *gin.Context) {
	type Params struct {
		ViewerId models.TwitchId `json:"viewer_id"`
		Username string          `json:"username"`
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName := ctx.Param(ChannelName)
	channelId, err := c.Users.GetUserId(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	viewer, err := c.Pets.GetPet(params.ViewerId, channelId, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceJoin(channelName, viewer)
}

func (c *TwitchBotController) RemoveViewerFromChannel(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	viewerId := models.TwitchId(ctx.Param(ViewerId))

	c.Announcer.AnnouncePart(channelName, viewerId)
}

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	action := ctx.Param(Action)
	viewerId := models.TwitchId(ctx.Param(ViewerId))

	c.Announcer.AnnounceAction(channelName, action, viewerId)
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
	viewerId := models.TwitchId(ctx.Param(ViewerId))

	channelId, err := c.Users.GetUserId(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.Items.GetItemByName(channelId, params.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Items.SetSelectedItem(viewerId, channelId, item.ItemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, viewerId)
}
