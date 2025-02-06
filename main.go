package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/routes"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

func run() error {
	env := os.Getenv("ENVIRONMENT")
	if env != "PRODUCTION" {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	db := config.ConnectDB()

	twitchApi := twitch.New(http.DefaultClient, "https://id.twitch.tv")
	itemRepo := repositories.NewItemRepository(db)
	channels := repositories.NewChannelRepo(db)

	auth := config.CreateAuthService(channels)

	announcer := announcers.NewAnnouncerService()
	cachedAnnouncer := announcers.NewCachedAnnouncerService(announcer)

	items := services.NewItemService(itemRepo)
	pets := services.NewPetService(items)

	extension := controllers.NewExtensionController(cachedAnnouncer, auth, items)
	twitchBot := controllers.NewTwitchBotController(cachedAnnouncer, items, pets)

	r := gin.Default()
	routes.RegisterRoutes(
		r,
		extension,
		twitchBot,
		twitchApi,
		channels,
		cachedAnnouncer,
		auth,
	)

	return r.Run()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
