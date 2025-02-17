package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/gorm"
	"github.com/streampets/backend/items"
	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/routes"
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

	db := gorm.NewDB(gorm.DSN())
	db.Open()

	itemRepo := gorm.NewItemRepository(db)
	channelRepo := gorm.NewChannelRepository(db)

	auth := config.CreateAuthService()

	twitchApi := twitch.New(http.DefaultClient, "https://id.twitch.tv")

	announcer := announcers.NewAnnouncer()
	cachedAnnouncer := announcers.NewCachedAnnouncer(announcer)

	items := items.New(itemRepo)
	pets := pets.New(items)

	r := gin.Default()
	routes.RegisterRoutes(
		r,
		channelRepo,
		twitchApi,
		cachedAnnouncer,
		auth,
		items,
		pets,
	)

	return r.Run()
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
