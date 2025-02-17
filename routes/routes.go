package routes

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/auth"
	"github.com/streampets/backend/gorm"
	"github.com/streampets/backend/items"
	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/twitch"
)

func RegisterRoutes(
	r *gin.Engine,
	channelRepo *gorm.ChannelRepository,
	twitch *twitch.TwitchApi,
	announcer *announcers.CachedAnnouncer,
	authService *auth.AuthService,
	store *items.ItemService,
	pets *pets.PetService,
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

	r.GET("/overlay/listen",
		auth.ListenAuthentication(channelRepo.GetOverlayId),
		handleListen(announcer.AddClient, announcer.RemoveClient),
	)

	extension := r.Group("/extension")
	{
		extension.Use(authService.ExtensionMiddleware())

		extension.GET("/items",
			handleGetStoreData(store.GetChannelsItems),
		)
		extension.GET("/user",
			handleGetUserData(store.GetSelectedItem, store.GetOwnedItems),
		)
		extension.POST("/items",
			authService.ReceiptMiddleware(),
			handleBuyStoreItem(store.GetItemById, store.AddOwnedItem),
		)
		extension.PUT("/items",
			handleSetSelectedItem(announcer.AnnounceUpdate, store.GetItemById, store.SetSelectedItem),
		)
	}

	r.GET("/dashboard/login",
		twitch.AuthorizationMiddleware(),
		handleLogin(channelRepo.GetOverlayId),
	)

	r.POST("/channels/:channelId/users",
		handleAddPetToChannel(announcer.AnnounceJoin, pets.GetPet),
	)
	r.DELETE("/channels/:channelId/users/:userId",
		handleRemoveUserFromChannel(announcer.AnnouncePart),
	)
	r.POST("/channels/:channelId/users/:userId/:action",
		handleAction(announcer.AnnounceAction),
	)
	r.PUT("/channels/:channelId/users/:userId",
		handleUpdate(announcer.AnnounceUpdate, store.GetItemByName, store.SetSelectedItem),
	)
}
