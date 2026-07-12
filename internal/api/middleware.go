package api

import (
	"fmt"
	"time"

	"chaosguard/internal/api/responses"
	"chaosguard/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CORSMiddleware sets up standard CORS headers for React dashboard integration
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware injects a unique X-Request-ID header into the context and headers
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Set("RequestID", reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)
		c.Next()
	}
}

// LoggerMiddleware records request stats using ChaosGuard's pkg/logger
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get("RequestID")

		logger.Debug("[API] Status: %d | Latency: %v | Method: %s | Path: %s | Query: %s | RequestID: %v",
			status, latency, c.Request.Method, path, query, reqID)
	}
}

// RecoveryMiddleware catches panics and returns a structured JSON error response
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(fmt.Errorf("%v", err), "[API] Panic recovered")
				c.AbortWithStatusJSON(500, responses.ErrorResponse{
					Success: false,
					Error:   "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
