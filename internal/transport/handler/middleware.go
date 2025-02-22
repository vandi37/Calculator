package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ContentType(next gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Add("Content-Type", "application/json")
		next(ctx)
	}
}

func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.String()
		}

		c.Next()

		latency := time.Since(start)

		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = "-"
		}

		statusCode := c.Writer.Status()
		method := c.Request.Method

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("method", method),
			zap.String("path", path),
		}

		switch {
		case statusCode >= http.StatusInternalServerError:
			logger.Error("request", fields...)
		default:
			logger.Info("request", fields...)
		}

		for _, err := range c.Errors {
			logger.Error("gin error", zap.Error(err))
		}

	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
