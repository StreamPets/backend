package twitch

import (
	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix/v2"
)

var EVENT_PATH = map[string]string{
	helix.EventSubTypeChannelChatMessage:     "/message",
	helix.EventSubTypeChannelFollow:          "/follow",
	helix.EventSubTypeChannelSubscription:    "/sub",
	helix.EventSubTypeChannelSubscriptionEnd: "/sub-end",
}

var client *helix.Client
var uri string
var channels map[string]*TwitchChannel

func Init(URI, clientId, appAccessToken, token string) {
	var err error = nil
	client, err = helix.NewClient(&helix.Options{
		ClientID:       clientId,
		AppAccessToken: appAccessToken,
	})
	if err != nil {
		panic(err)
	}
	uri = URI + "/wh"
	channels = make(map[string]*TwitchChannel)
}

func Close() {
	for _, channel := range channels {
		channel.close()
	}
}

/******************************************************************************
       Following callback methods distributes events to dedicated channels
*****************************************************************************/

func OnMessageReceived(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubSubscription            `json:"subscription"`
		Event        helix.EventSubChannelChatMessageEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.Condition.BroadcasterUserID]; ok {
		ch.onMessageReceived(&msgEvent.Event)
	}
	ctx.String(200, "OK")
}

func OnFollow(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubSubscription       `json:"subscription"`
		Event        helix.EventSubChannelFollowEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.Condition.BroadcasterUserID]; ok {
		ch.onFollow(&msgEvent.Event)
	}
	ctx.String(200, "OK")
}

func OnSubscription(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubSubscription          `json:"subscription"`
		Event        helix.EventSubChannelSubscribeEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.Condition.BroadcasterUserID]; ok {
		ch.onSubscription(&msgEvent.Event, true)
	}
	ctx.String(200, "OK")
}

func OnSubscriptionEnd(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubSubscription          `json:"subscription"`
		Event        helix.EventSubChannelSubscribeEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.Condition.BroadcasterUserID]; ok {
		ch.onSubscription(&msgEvent.Event, false)
	}
	ctx.String(200, "OK")
}

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(403, gin.H{
		"message": err.Error(),
	})
}
