package routes

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/auth"
	"github.com/streampets/backend/database"
	"github.com/streampets/backend/items"
	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/twitch"
)

func RegisterRoutes(
	r *gin.Engine,
	db *database.DB,
	twitchApi *twitch.TwitchApi,
	announcer *announcers.CachedAnnouncerService,
	auth *auth.AuthService,
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
		handleListen(announcer.AddClient, announcer.RemoveClient, auth.ValidateOverlayId),
	)

	extension := r.Group("/extension")
	{
		extension.Use(
			auth.ExtensionMiddleware(),
		)
		extension.GET("/items",
			handleGetStoreData(store.GetChannelsItems),
		)
		extension.GET("/user",
			handleGetUserData(store.GetSelectedItem, store.GetOwnedItems),
		)
		extension.POST("/items",
			auth.ReceiptMiddleware(),
			handleBuyStoreItem(store.GetItemById, store.AddOwnedItem),
		)
		extension.PUT("/items",
			handleSetSelectedItem(announcer.AnnounceUpdate, store.GetItemById, store.SetSelectedItem),
		)
	}

	r.GET("/dashboard/login",
		handleLogin(twitchApi.ValidateToken, db.GetOverlayId),
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
