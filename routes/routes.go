package routes

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

func RegisterRoutes(
	r *gin.Engine,
	twitchBot *controllers.TwitchBotController,
	twitchApi *twitch.TwitchApi,
	channelRepo *repositories.ChannelRepo,
	announcer *announcers.CachedAnnouncerService,
	auth *services.AuthService,
	store *services.ItemService,
	pets *services.PetService,
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

	r.GET("/overlay/listen", handleListen(announcer, auth))

	r.GET("/extension/items", handleGetStoreData(auth, store))
	r.GET("/extension/user", handleGetUserData(auth, store))
	r.POST("/extension/items", handleBuyStoreItem(auth, store))
	r.PUT("/extension/items", handleSetSelectedItem(announcer, auth, store))

	r.GET("/dashboard/login", handleLogin(twitchApi, channelRepo))

	r.POST("/channels/:channelId/users", handleAddPetToChannel(announcer, pets))
	r.DELETE("/channels/:channelId/users/:userId", handleRemoveUserFromChannel(announcer))
	r.POST("/channels/:channelId/users/:userId/:action", handleAction(announcer))
	r.PUT("/channels/:channelId/users/:userId", twitchBot.UpdateUser)
}
