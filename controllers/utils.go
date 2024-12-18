package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const XExtensionJwt string = "x-extension-jwt"
const Action string = "action"

const ChannelID string = "channelID"
const ChannelName string = "channelName"
const OverlayID string = "overlayID"
const UserID string = "userID"

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
