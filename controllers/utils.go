package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const XExtensionJwt string = "x-extension-jwt"
const Action string = "action"

const ChannelId string = "channelId"
const ChannelName string = "channelName"
const OverlayId string = "overlayId"
const ViewerId string = "viewerId"

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
