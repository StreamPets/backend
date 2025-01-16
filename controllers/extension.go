package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type UpdateAnnouncer interface {
	AnnounceUpdate(channelName, image string, viewerId models.TwitchId)
}

type TokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
	VerifyReceipt(tokenString string) (*services.Receipt, error)
}

type StoreService interface {
	GetItemById(itemId uuid.UUID) (models.Item, error)
	GetSelectedItem(viewerId, channelId models.TwitchId) (models.Item, error)
	SetSelectedItem(viewerId, channelId models.TwitchId, itemId uuid.UUID) error
	GetChannelsItems(channelId models.TwitchId) ([]models.Item, error)
	GetOwnedItems(channelId, viewerId models.TwitchId) ([]models.Item, error)
	AddOwnedItem(viewerId models.TwitchId, itemId, transactionId uuid.UUID) error
}

type ExtensionController struct {
	Announcer UpdateAnnouncer
	Verifier  TokenVerifier
	Store     StoreService
	Usernames UsernameGetter
}

func NewExtensionController(
	announcer UpdateAnnouncer,
	verifier TokenVerifier,
	store StoreService,
	usernames UsernameGetter,
) *ExtensionController {
	return &ExtensionController{
		Announcer: announcer,
		Verifier:  verifier,
		Store:     store,
		Usernames: usernames,
	}
}

func (c *ExtensionController) GetStoreData(ctx *gin.Context) {
	tokenString := ctx.GetHeader(XExtensionJwt)

	token, err := c.Verifier.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	storeItems, err := c.Store.GetChannelsItems(token.ChannelId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, storeItems)
}

func (c *ExtensionController) GetViewerData(ctx *gin.Context) {
	tokenString := ctx.GetHeader(XExtensionJwt)

	token, err := c.Verifier.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ownedItems, err := c.Store.GetOwnedItems(token.ChannelId, token.ViewerId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	selectedItem, err := c.Store.GetSelectedItem(token.ViewerId, token.ChannelId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"selected": selectedItem,
		"owned":    ownedItems,
	})
}

func (c *ExtensionController) BuyStoreItem(ctx *gin.Context) {
	type Params struct {
		Receipt string `json:"receipt"`
		ItemId  string `json:"item_id"`
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

	receipt, err := c.Verifier.VerifyReceipt(params.Receipt)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.Store.AddOwnedItem(token.ViewerId, itemId, receipt.TransactionId); err != nil {
		addErrorToCtx(err, ctx)
		return
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

	if err = c.Store.SetSelectedItem(token.ViewerId, token.ChannelId, itemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.Usernames.GetUsername(token.ChannelId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, token.ViewerId)
}
