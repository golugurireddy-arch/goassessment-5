package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		reqID, _ := c.Get(HeaderXRequestID)
		errStr := c.Errors.ByType(gin.ErrorTypePrivate).String()

		log.Printf("request_id=%v method=%s path=%s status=%d duration=%v error=%s",
			reqID, c.Request.Method, path, status, duration, errStr,
		)
	}
}
