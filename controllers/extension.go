package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type UpdateAnnouncer interface {
	AnnounceUpdate(channelName, image string, userID models.TwitchID)
}

type TokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
	VerifyReceipt(tokenString string) (*services.Receipt, error)
}

type StoreService interface {
	GetItemByID(itemID uuid.UUID) (models.Item, error)
	SetSelectedItem(userID, channelID models.TwitchID, itemID uuid.UUID) error
	GetTodaysItems(channelID models.TwitchID) ([]models.Item, error)
	GetOwnedItems(channelID, userID models.TwitchID) ([]models.Item, error)
	AddOwnedItem(userID models.TwitchID, itemID, transactionID uuid.UUID) error
}

type UserGetter interface {
	GetUsername(userID models.TwitchID) (string, error)
	GetUserID(username string) (models.TwitchID, error)
}

type ExtensionController struct {
	Announcer UpdateAnnouncer
	Verifier  TokenVerifier
	Store     StoreService
	Users     UserGetter
}

func NewExtensionController(
	announcer UpdateAnnouncer,
	verifier TokenVerifier,
	store StoreService,
	users UserGetter,
) *ExtensionController {
	return &ExtensionController{
		Announcer: announcer,
		Verifier:  verifier,
		Store:     store,
		Users:     users,
	}
}

func (c *ExtensionController) GetStoreData(ctx *gin.Context) {
	tokenString := ctx.GetHeader(XExtensionJwt)

	token, err := c.Verifier.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	storeItems, err := c.Store.GetTodaysItems(token.ChannelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ownedItems, err := c.Store.GetOwnedItems(token.ChannelID, token.UserID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"store": storeItems,
		"owned": ownedItems,
	})
}

func (c *ExtensionController) BuyStoreItem(ctx *gin.Context) {
	type Params struct {
		Receipt string `json:"receipt"`
		ItemID  string `json:"item_id"`
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

	itemID, err := uuid.Parse(params.ItemID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	receipt, err := c.Verifier.VerifyReceipt(params.Receipt)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.Store.AddOwnedItem(token.UserID, itemID, receipt.TransactionID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
}

func (c *ExtensionController) SetSelectedItem(ctx *gin.Context) {
	type Params struct {
		ItemID string `json:"item_id"`
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

	itemID, err := uuid.Parse(params.ItemID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.Store.GetItemByID(itemID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.Store.SetSelectedItem(token.UserID, token.ChannelID, itemID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.Users.GetUsername(token.ChannelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.Announcer.AnnounceUpdate(channelName, item.Image, token.UserID)
}
