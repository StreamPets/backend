package twitch

import (
	"github.com/gin-gonic/gin"
	"github.com/nicklaw5/helix/v2"
)

func OnStreamStarted(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubStreamOnlineEvent       `json:"subscription"`
		Event        helix.EventSubChannelChatMessageEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.BroadcasterUserID]; ok {
		ch.onStreamOnline(&msgEvent.Event)
	}
	ctx.String(200, "OK")
}

func OnStreamStopped(ctx *gin.Context) {
	type eventMessage struct {
		Subscription helix.EventSubStreamOfflineEvent      `json:"subscription"`
		Event        helix.EventSubChannelChatMessageEvent `json:"event"`
	}
	var msgEvent eventMessage
	if err := ctx.ShouldBindJSON(&msgEvent); err != nil {
		addErrorToCtx(err, ctx)
		return
	}
	if ch, ok := channels[msgEvent.Subscription.BroadcasterUserID]; ok {
		ch.onStreamOffline(&msgEvent.Event)
	}
	ctx.String(200, "OK")
}
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
