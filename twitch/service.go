package twitch

import "github.com/gin-gonic/gin"

type TwitchController struct {
}

func (c *TwitchController) OnMessageReceived(ctx *gin.Context) {
}

func (c *TwitchController) OnFollow(ctx *gin.Context) {
}

func (c *TwitchController) OnBanEnabled(ctx *gin.Context) {
}

func (c *TwitchController) OnBanDisabled(ctx *gin.Context) {
}

func (c *TwitchController) OnSubscription(ctx *gin.Context) {
}

func (c *TwitchController) OnSubscriptionEnd(ctx *gin.Context) {
}
