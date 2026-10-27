package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole checks whether the authenticated user has any of the permitted roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(string(RoleKey))
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: No role assigned"})
			return
		}

		userRole, _ := roleVal.(string)
		for _, allowed := range roles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient privileges"})
	}
}
