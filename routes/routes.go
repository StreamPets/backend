package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
	"gorm.io/gorm"
)

func RegisterRoutes(
	r *gin.Engine,
	db *gorm.DB,
) {
	twitchApi := twitch.NewTwitchApi(http.DefaultClient)
	itemRepo := repositories.NewItemRepository(db)
	channels := repositories.NewChannelRepo(db)

	auth := config.CreateAuthService(channels)
	cache := services.NewPetCacheService()
	announcer := services.NewAnnouncerService(cache)
	items := services.NewItemService(itemRepo)
	pets := services.NewPetService(items)

	overlay := controllers.NewOverlayController(announcer, auth)
	extension := controllers.NewExtensionController(announcer, auth, items)
	dashboard := controllers.NewDashboardController(channels.GetOverlayId, twitchApi.ValidateToken)

	twitchBot := controllers.NewTwitchBotController(announcer, items, pets)

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

	dashRouter := r.Group("/dashboard")
	{
		dashRouter.GET("/login", dashboard.HandleLogin)
	}

	api := r.Group("/channels")
	{
		api.POST("/:channelId/users", twitchBot.AddPetToChannel)
		api.DELETE("/:channelId/users/:userId", twitchBot.RemoveUserFromChannel)
		api.POST("/:channelId/users/:userId/:action", twitchBot.Action)
		api.PUT("/:channelId/users/:userId", twitchBot.UpdateUser)
	}
}
