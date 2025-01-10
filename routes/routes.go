package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
)

func RegisterRoutes(
	r *gin.Engine,
	db *gorm.DB,
	twitchRepo *repositories.TwitchRepository,
	authService *services.AuthService,
) {
	itemRepo := repositories.NewItemRepository(db)

	announcer := services.NewAnnouncerService()
	items := services.NewItemService(itemRepo)
	viewers := services.NewViewerService(itemRepo)
	cache := services.NewViewerCacheService()

	overlay := controllers.NewOverlayController(announcer, authService, twitchRepo, cache)
	twitchBot := controllers.NewTwitchBotController(announcer, items, viewers, twitchRepo, cache)
	extension := controllers.NewExtensionController(announcer, authService, items, twitchRepo)

	api := r.Group("/overlay")
	{
		api.GET("/listen", overlay.HandleListen)
	}

	api = r.Group("/channels")
	{
		api.POST("/:channelName/viewers", twitchBot.AddViewerToChannel)
		api.DELETE("/:channelName/viewers/:userID", twitchBot.RemoveViewerFromChannel)
		api.POST("/:channelName/viewers/:userID/:action", twitchBot.Action)
		api.PUT("/:channelName/viewers/:userID", twitchBot.UpdateViewer)
	}

	api = r.Group("/items")
	{
		api.GET("/", extension.GetStoreData)
		api.POST("/", extension.BuyStoreItem)
		api.PUT("/", extension.SetSelectedItem)
	}

	api = r.Group("/wh")
	{
		api.POST("/message", twitch.OnMessageReceived)
		api.POST("/follow", twitch.OnFollow)
		api.POST("/ban", twitch.OnBanEnabled)
		api.POST("/ban-end", twitch.OnBanDisabled)
		api.POST("/sub", twitch.OnSubscription)
		api.POST("/sub-end", twitch.OnSubscriptionEnd)
	}
}
