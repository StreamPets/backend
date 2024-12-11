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
	databaseService := services.NewDatabaseService(itemRepo, twitchRepo)

	controller := controllers.NewController(
		announcer,
		authService,
		twitchRepo,
		databaseService,
	)

	api := r.Group("/overlay")
	{
		api.GET("/listen", controller.HandleListen)
	}

	api = r.Group("/channels")
	{
		api.POST("/:channelName/viewers", controller.AddViewerToChannel)
		api.DELETE("/:channelName/viewers/:userID", controller.RemoveViewerFromChannel)
		api.POST("/:channelName/viewers/:userID/:action", controller.Action)
		api.PUT("/:channelName/viewers/:userID", controller.UpdateViewer)
	}

	api = r.Group("/items")
	{
		api.GET("/", controller.GetStoreData)
		api.POST("/", controller.BuyStoreItem)
		api.PUT("/", controller.SetSelectedItem)
	}
}
