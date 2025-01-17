package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type Announcer interface {
	AnnounceJoin(channelName string, pet services.Pet)
	AnnouncePart(channelName string, userId models.TwitchId)
	AnnounceAction(channelName, action string, userId models.TwitchId)
	AnnounceUpdate(channelName, image string, userId models.TwitchId)
}

type PetGetter interface {
	GetPet(userId, channelId models.TwitchId, username string) (services.Pet, error)
}

type ItemGetSetter interface {
	GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error)
	SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error
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

func (c *TwitchBotController) AddPetToChannel(ctx *gin.Context) {
	type Params struct {
		UserId   models.TwitchId `json:"user_id"`
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

	pet, err := c.Pets.GetPet(params.UserId, channelId, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceJoin(channelName, pet)
}

func (c *TwitchBotController) RemoveUserFromChannel(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	userId := models.TwitchId(ctx.Param(UserId))

	c.Announcer.AnnouncePart(channelName, userId)
}

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelName := ctx.Param(ChannelName)
	action := ctx.Param(Action)
	userId := models.TwitchId(ctx.Param(UserId))

	c.Announcer.AnnounceAction(channelName, action, userId)
}

func (c *TwitchBotController) UpdateUser(ctx *gin.Context) {
	type Params struct {
		ItemName string `json:"item_name"`
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName := ctx.Param(ChannelName)
	userId := models.TwitchId(ctx.Param(UserId))

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

	if err = c.Items.SetSelectedItem(userId, channelId, item.ItemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, userId)
}
