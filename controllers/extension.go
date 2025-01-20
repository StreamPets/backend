package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type UpdateAnnouncer interface {
	AnnounceUpdate(channelName, image string, userId models.TwitchId)
}

type TokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
	VerifyReceipt(receiptString string) (*services.Receipt, error)
}

type StoreService interface {
	GetItemById(itemId uuid.UUID) (models.Item, error)
	GetSelectedItem(userId, channelId models.TwitchId) (models.Item, error)
	SetSelectedItem(userId, channelId models.TwitchId, itemId uuid.UUID) error
	GetChannelsItems(channelId models.TwitchId) ([]models.Item, error)
	GetOwnedItems(channelId, userId models.TwitchId) ([]models.Item, error)
	AddOwnedItem(userId models.TwitchId, itemId, transactionId uuid.UUID) error
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

// Invalid channel id or invalid user id
// Invalid item (unowned)

func (c *ExtensionController) GetUserData(ctx *gin.Context) {
	tokenString := ctx.GetHeader(XExtensionJwt)

	token, err := c.Verifier.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ownedItems, err := c.Store.GetOwnedItems(token.ChannelId, token.UserId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	selectedItem, err := c.Store.GetSelectedItem(token.UserId, token.ChannelId)
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

	receipt, err := c.Verifier.VerifyReceipt(params.Receipt)
	if err != nil {
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

	if item.Rarity != receipt.Data.Product.Rarity {
		addErrorToCtx(errors.New("receipt and item rarity do not match"), ctx)
		return
	}

	if err := c.Store.AddOwnedItem(token.UserId, itemId, receipt.Data.TransactionId); err != nil {
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

	if err = c.Store.SetSelectedItem(token.UserId, token.ChannelId, itemId); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.Usernames.GetUsername(token.ChannelId)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, token.UserId)
}
