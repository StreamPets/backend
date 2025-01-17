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

	cache := services.NewPetCacheService()
	announcer := services.NewAnnouncerService(cache)
	items := services.NewItemService(itemRepo)
	petService := services.NewPetService(itemRepo)

	overlay := controllers.NewOverlayController(announcer, authService, twitchRepo)
	extension := controllers.NewExtensionController(announcer, authService, items, twitchRepo)

	twitchBot := controllers.NewTwitchBotController(announcer, items, petService, twitchRepo)

	overlayRouter := r.Group("/overlay")
	{
		overlayRouter.GET("/listen", overlay.HandleListen)
	}

	extRouter := r.Group("/extension")
	{
		extRouter.GET("/viewer", extension.GetViewerData)
		extRouter.GET("/items", extension.GetStoreData)
		extRouter.POST("/items", extension.BuyStoreItem)
		extRouter.PUT("/items", extension.SetSelectedItem)
	}

	api := r.Group("/channels")
	{
		api.POST("/:channelName/viewers", twitchBot.AddViewerToChannel)
		api.DELETE("/:channelName/viewers/:viewerId", twitchBot.RemoveViewerFromChannel)
		api.POST("/:channelName/viewers/:viewerId/:action", twitchBot.Action)
		api.PUT("/:channelName/viewers/:viewerId", twitchBot.UpdateViewer)
	}
}
