package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

type Controller interface {
	HandleListen(ctx *gin.Context)

	AddViewerToChannel(ctx *gin.Context)
	RemoveViewerFromChannel(ctx *gin.Context)
	Action(ctx *gin.Context)
	UpdateViewer(ctx *gin.Context)

	GetStoreData(ctx *gin.Context)
	BuyStoreItem(ctx *gin.Context)
	SetSelectedItem(ctx *gin.Context)
}

type controller struct {
	announcer       services.Announcer
	authService     services.AuthService
	twitchRepo      repositories.TwitchRepository
	databaseService services.DatabaseService
}

func NewController(
	announcer services.Announcer,
	authService services.AuthService,
	twitchRepo repositories.TwitchRepository,
	databaseService services.DatabaseService,
) Controller {
	return &controller{
		announcer:       announcer,
		authService:     authService,
		twitchRepo:      twitchRepo,
		databaseService: databaseService,
	}
}

func (c *controller) HandleListen(ctx *gin.Context) {
	channelID := models.TwitchID(ctx.Query("channelID"))
	overlayID, err := uuid.Parse(ctx.Query("overlayID"))
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.authService.VerifyOverlayID(channelID, overlayID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.twitchRepo.GetUsername(channelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	client := c.announcer.AddClient(channelName)
	defer func() {
		go func() {
			for range client.Stream {
			}
		}()
		c.announcer.RemoveClient(client)
	}()

	ctx.Stream(func(w io.Writer) bool {
		if event, ok := <-client.Stream; ok {
			ctx.SSEvent(event.Event, event.Message)
			return true
		}
		return false
	})
}

func (c *controller) AddViewerToChannel(ctx *gin.Context) {
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

	viewer, err := c.databaseService.GetViewer(params.UserID, channelName, params.Username)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.announcer.AnnounceJoin(channelName, viewer)
}

func (c *controller) RemoveViewerFromChannel(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	userID := models.TwitchID(ctx.Param("userID"))

	c.announcer.AnnouncePart(channelName, userID)
}

func (c *controller) Action(ctx *gin.Context) {
	channelName := ctx.Param("channelName")
	action := ctx.Param("action")
	userID := models.TwitchID(ctx.Param("userID"))

	c.announcer.AnnounceAction(channelName, action, userID)
}

func (c *controller) UpdateViewer(ctx *gin.Context) {
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

	channelID, err := c.twitchRepo.GetUserID(channelName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	item, err := c.databaseService.GetItemByName(channelID, params.ItemName)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.databaseService.SetSelectedItem(userID, channelID, item.ItemID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.announcer.AnnounceUpdate(channelName, item.Image, userID)
}

func (c *controller) GetStoreData(ctx *gin.Context) {
	tokenString := ctx.GetHeader("x-extension-jwt")

	token, err := c.authService.VerifyExtToken(tokenString)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	storeItems, err := c.databaseService.GetTodaysItems(token.ChannelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ownedItems, err := c.databaseService.GetOwnedItems(token.ChannelID, token.UserID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"store": storeItems,
		"owned": ownedItems,
	})
}

func (c *controller) BuyStoreItem(ctx *gin.Context) {
	type Params struct {
		Receipt string `json:"receipt"`
		ItemID  string `json:"item_id"`
	}

	tokenString := ctx.GetHeader("x-extension-jwt")
	token, err := c.authService.VerifyExtToken(tokenString)
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

	receipt, err := c.authService.VerifyReceipt(params.Receipt)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err := c.databaseService.AddOwnedItem(token.UserID, itemID, receipt.TransactionID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
}

func (c *controller) SetSelectedItem(ctx *gin.Context) {
	type Params struct {
		ItemID string `json:"item_id"`
	}

	tokenString := ctx.GetHeader("x-extension-jwt")
	token, err := c.authService.VerifyExtToken(tokenString)
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

	item, err := c.databaseService.GetItemByID(itemID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	if err = c.databaseService.SetSelectedItem(token.UserID, token.ChannelID, itemID); err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	channelName, err := c.twitchRepo.GetUsername(token.ChannelID)
	if err != nil {
		addErrorToCtx(err, ctx)
		return
	}

	c.announcer.AnnounceUpdate(channelName, item.Image, token.UserID)
}

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
