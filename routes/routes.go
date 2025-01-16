package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
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
	extension := controllers.NewExtensionController(announcer, authService, items, twitchRepo)

	twitchBot := controllers.NewTwitchBotController(announcer, items, viewers, twitchRepo, cache)

	overlayRouter := r.Group("/overlay")
	{
		overlayRouter.GET("/listen", overlay.HandleListen)
	}

	extRouter := r.Group("/extension")
	{
		extRouter.GET("/user", extension.GetUserData)
		extRouter.GET("/items", extension.GetStoreData)
		extRouter.POST("/items", extension.BuyStoreItem)
		extRouter.PUT("/items", extension.SetSelectedItem)
	}

	api := r.Group("/channels")
	{
		api.POST("/:channelName/viewers", twitchBot.AddViewerToChannel)
		api.DELETE("/:channelName/viewers/:userID", twitchBot.RemoveViewerFromChannel)
		api.POST("/:channelName/viewers/:userID/:action", twitchBot.Action)
		api.PUT("/:channelName/viewers/:userID", twitchBot.UpdateViewer)
	}
}
