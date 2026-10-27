package middleware

import (
	"net/http"

	"order-service/repository"

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

// RequireOrderAccess verifies that the user is an ADMIN or owns the target order
func RequireOrderAccess(repo repository.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		if orderID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Order ID parameter missing"})
			return
		}

		role, _ := c.Get(string(RoleKey))
		if role == "ADMIN" {
			c.Next()
			return
		}

		userVal, _ := c.Get(string(UserKey))
		currentUserID, _ := userVal.(string)

		order, err := repo.GetByID(c.Request.Context(), orderID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		if order.CustomerID != currentUserID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Access denied to this order"})
			return
		}

		c.Set("current_order", order)
		c.Next()
	}
}
