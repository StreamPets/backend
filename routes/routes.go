package routes

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/controllers"
)

func RegisterRoutes(
	r *gin.Engine,
	overlay *controllers.OverlayController,
	extension *controllers.ExtensionController,
	dashboard *controllers.DashboardController,
	twitchBot *controllers.TwitchBotController,
) {
	overlayUrl := os.Getenv("OVERLAY_URL")
	extensionUrl := os.Getenv("EXTENSION_URL")
	dashboardUrl := os.Getenv("DASHBOARD_URL")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{overlayUrl, extensionUrl, dashboardUrl},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	r.GET("/overlay/listen", overlay.HandleListen)

	r.GET("/extension/user", extension.GetUserData)
	r.GET("/extension/items", extension.GetStoreData)
	r.POST("/extension/items", extension.BuyStoreItem)
	r.PUT("/extension/items", extension.SetSelectedItem)

	r.GET("/dashboard/login", dashboard.HandleLogin)

	r.POST("/channels/:channelId/users", twitchBot.AddPetToChannel)
	r.DELETE("/channels/:channelId/users/:userId", twitchBot.RemoveUserFromChannel)
	r.POST("/channels/:channelId/users/:userId/:action", twitchBot.Action)
	r.PUT("/channels/:channelId/users/:userId", twitchBot.UpdateUser)
}
