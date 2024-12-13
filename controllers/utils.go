package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func addErrorToCtx(err error, ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"message": err.Error(),
	})
}
