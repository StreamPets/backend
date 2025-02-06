package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

type Announcer interface {
	AnnounceAction(channelId, userId twitch.Id, action string)
	AnnounceUpdate(channelId, userId twitch.Id, image string)
}

type PetGetter interface {
	GetPet(userId, channelId twitch.Id, username string) (services.Pet, error)
}

type ItemGetSetter interface {
	GetItemByName(channelId twitch.Id, itemName string) (models.Item, error)
	SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error
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

func (c *TwitchBotController) Action(ctx *gin.Context) {
	channelId := twitch.Id(ctx.Param(ChannelId))
	userId := twitch.Id(ctx.Param(UserId))
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

	channelId := twitch.Id(ctx.Param(ChannelId))
	userId := twitch.Id(ctx.Param(UserId))

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
