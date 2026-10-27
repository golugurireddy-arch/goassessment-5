package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Get(HeaderXRequestID)
				log.Printf("request_id=%v panic recovered: %v", reqID, r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error occurred",
				})
			}
		}()
		c.Next()
	}
}
