package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streampets/backend/config"
	"github.com/streampets/backend/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	db := config.ConnectDB()
	twitch := config.CreateTwitchRepo()
	auth := config.CreateAuthService(db)

	r := gin.Default()

	routes.RegisterRoutes(r, db, twitch, auth)

	r.Run()
}
