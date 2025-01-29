package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const XExtensionJwt string = "x-extension-jwt"
const Action string = "action"

const ChannelId string = "channelId"
const OverlayId string = "overlayId"
const UserId string = "userId"

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
