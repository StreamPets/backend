package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/database"
	"github.com/streampets/backend/log"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/routes"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

func run(ctx context.Context, cfg *config.AppConfig, logger *slog.Logger) error {
	db, err := database.ConnectDB(cfg.Database)
	if err != nil {
		return err
	}

	twitchApi := twitch.New(http.DefaultClient, "https://id.twitch.tv")
	itemRepo := repositories.NewItemRepository(db)
	channels := repositories.NewChannelRepo(db)


	auth := services.NewAuthService(channels, cfg.ExtensionSecret)

	announcer := announcers.NewAnnouncerService()
	cachedAnnouncer := announcers.NewCachedAnnouncerService(announcer)

	items := services.NewItemService(itemRepo)
	pets := services.NewPetService(items)

	overlay := controllers.NewOverlayController(cachedAnnouncer, auth)
	extension := controllers.NewExtensionController(cachedAnnouncer, auth, items)
	dashboard := controllers.NewDashboardController(channels, twitchApi)
	twitchBot := controllers.NewTwitchBotController(cachedAnnouncer, items, pets)

	r := gin.Default()
	routes.RegisterRoutes(r, overlay, extension, dashboard, twitchBot)

	logger.InfoContext(ctx, "starting werbserver", "port", cfg.HttpPort)

	return r.Run()
}

func main() {
	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	logger := log.New(
		log.WithLevel(cfg.LogLevel),
		log.WithSource(), // Show where the log happend, file and line
	)

	if err := run(ctx, cfg, logger); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
