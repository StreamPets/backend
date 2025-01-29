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
	authService *services.AuthService,
) {
	itemRepo := repositories.NewItemRepository(db)

	cache := services.NewPetCacheService()
	announcer := services.NewAnnouncerService(cache)
	items := services.NewItemService(itemRepo)
	petService := services.NewPetService(items)

	overlay := controllers.NewOverlayController(announcer, authService)
	extension := controllers.NewExtensionController(announcer, authService, items)

	twitchBot := controllers.NewTwitchBotController(announcer, items, petService)

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
		api.POST("/:channelId/users", twitchBot.AddPetToChannel)
		api.DELETE("/:channelId/users/:userId", twitchBot.RemoveUserFromChannel)
		api.POST("/:channelId/users/:userId/:action", twitchBot.Action)
		api.PUT("/:channelId/users/:userId", twitchBot.UpdateUser)
	}
}
