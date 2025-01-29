package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type Announcer interface {
	AnnounceJoin(channelId models.TwitchId, pet services.Pet)
	AnnouncePart(channelId, userId models.TwitchId)
	AnnounceAction(channelId, userId models.TwitchId, action string)
	AnnounceUpdate(channelId, userId models.TwitchId, image string)
}

type PetGetter interface {
	GetPet(userId, channelId models.TwitchId, username string) (services.Pet, error)
}

type ItemGetSetter interface {
	GetItemByName(channelId models.TwitchId, itemName string) (models.Item, error)
	SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error
}

type TwitchBotController struct {
	Announcer Announcer
	Items     ItemGetSetter
	Pets      PetGetter
}

func NewTwitchBotController(
	announcer Announcer,
	items ItemGetSetter,
	pets PetGetter,
) *TwitchBotController {
	return &TwitchBotController{
		Announcer: announcer,
		Items:     items,
		Pets:      pets,
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

	channelId := models.TwitchId(ctx.Param(ChannelId))
	pet, err := c.Pets.GetPet(params.UserId, channelId, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceJoin(channelId, pet)
	ctx.JSON(http.StatusNoContent, nil)
}

func (c *TwitchBotController) RemoveUserFromChannel(ctx *gin.Context) {
	channelId := models.TwitchId(ctx.Param(ChannelId))
	userId := models.TwitchId(ctx.Param(UserId))

	c.Announcer.AnnouncePart(channelId, userId)
	ctx.JSON(http.StatusNoContent, nil)
}

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelId := models.TwitchId(ctx.Param(ChannelId))
	userId := models.TwitchId(ctx.Param(UserId))
	action := ctx.Param(Action)

	c.Announcer.AnnounceAction(channelId, userId, action)
	ctx.JSON(http.StatusNoContent, nil)
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

	channelId := models.TwitchId(ctx.Param(ChannelId))
	userId := models.TwitchId(ctx.Param(UserId))

	item, err := c.Items.GetItemByName(channelId, params.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Items.SetSelectedItem(userId, channelId, item.ItemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelId, userId, item.Image)
	ctx.JSON(http.StatusNoContent, nil)
}
