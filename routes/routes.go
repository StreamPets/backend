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
	twitchRepo repositories.TwitchRepository,
) {
	channelRepo := repositories.NewChannelRepo(db)
	itemRepo := repositories.NewItemRepository(db)

	announcer := services.NewAnnounceService()
	authService := services.NewAuthService(channelRepo, "")
	viewerService := services.NewViewerService(itemRepo, twitchRepo)

	controller := controllers.NewController(
		announcer,
		authService,
		twitchRepo,
		viewerService,
	)

	api := r.Group("/overlay")
	{
		api.GET("/listen", controller.HandleListen)
	}

	api = r.Group("/channels")
	{
		api.POST("/:channelID/viewers", controller.AddViewerToChannel)
		api.DELETE("/:channelID/viewers/:userID", controller.RemoveViewerFromChannel)
		api.POST("/:channelID/viewers/:userID/:action", controller.Action)
		api.PUT("/:channelID/viewers/:userID", controller.UpdateViewer)
	}

	api = r.Group("/items")
	{
		api.GET("/", controller.GetStoreData)
	}
}
