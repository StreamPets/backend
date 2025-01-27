package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/routes"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env != "PRODUCTION" {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}
	}

	db := config.ConnectDB()
	twitch := config.CreateTwitchRepo()
	auth := config.CreateAuthService(db)

	r := gin.Default()

	// overlayUrl := os.Getenv("OVERLAY_URL")
	// extensionUrl := os.Getenv("EXTENSION_URL")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	routes.RegisterRoutes(r, db, twitch, auth)
	r.Run()
}
