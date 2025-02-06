package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

type UpdateAnnouncer interface {
	AnnounceUpdate(channelId, userId twitch.Id, image string)
}

type tokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
	VerifyReceipt(receiptString string) (*services.Receipt, error)
}

type StoreService interface {
	GetItemById(itemId uuid.UUID) (models.Item, error)
	GetSelectedItem(userId, channelId twitch.Id) (models.Item, error)
	SetSelectedItem(userId, channelId twitch.Id, itemId uuid.UUID) error
	GetChannelsItems(channelId twitch.Id) ([]models.Item, error)
	GetOwnedItems(channelId, userId twitch.Id) ([]models.Item, error)
	AddOwnedItem(userId twitch.Id, itemId, transactionId uuid.UUID) error
}

type ExtensionController struct {
	Announcer UpdateAnnouncer
	Verifier  tokenVerifier
	Store     StoreService
}

func NewExtensionController(
	announcer UpdateAnnouncer,
	verifier tokenVerifier,
	store StoreService,
) *ExtensionController {
	return &ExtensionController{
		Announcer: announcer,
		Verifier:  verifier,
		Store:     store,
	}
}

func (c *ExtensionController) SetSelectedItem(ctx *gin.Context) {
	type Params struct {
		ItemId string `json:"item_id"`
	}

	tokenString := ctx.GetHeader(XExtensionJwt)
	token, err := c.Verifier.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	var params Params
	if err := ctx.ShouldBindJSON(&params); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	itemId, err := uuid.Parse(params.ItemId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.Store.GetItemById(itemId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Store.SetSelectedItem(token.UserId, token.ChannelId, itemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(token.ChannelId, token.UserId, item.Image)
}
