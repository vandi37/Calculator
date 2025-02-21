package handler

import (
	"github.com/gin-gonic/gin"
)

func ContentType(next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Add("Content-Type", "application/json")
		next(ctx)
	}
}
