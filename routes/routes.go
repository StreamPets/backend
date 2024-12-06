package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, twitchRepo repositories.Twitcher) {
	channelRepo := repositories.NewChannelRepository(db)
	itemRepo := repositories.NewItemRepository(db)

	announcer := services.NewAnnounceService()
	authService := services.NewAuthService(channelRepo)
	viewerService := services.NewViewerService(itemRepo, twitchRepo)

	overlayController := controllers.NewOverlayController(
		announcer,
		authService,
		twitchRepo,
		viewerService,
	)

	api := r.Group("/overlay")
	{
		api.GET("/listen", overlayController.HandleListen)
	}

	api = r.Group("/channels")
	{
		api.POST("/:channelID/viewers", overlayController.AddViewerToChannel)
		api.DELETE("/:channelID/viewers/:userID", overlayController.RemoveViewerFromChannel)
		api.POST("/:channelID/viewers/:userID/:action", overlayController.Action)
		api.PUT("/:channelID/viewers/:userID", overlayController.UpdateViewer)
	}
}
