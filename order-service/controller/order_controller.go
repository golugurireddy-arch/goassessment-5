package controller

import (
	"errors"
	"net/http"

	"order-service/model"
	"order-service/service"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	svc service.OrderService
}

func NewOrderController(svc service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

// POST /api/v1/orders
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var dto model.CreateOrderDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If customerId is not sent in the body, fallback to authenticated user ID
	customerID := dto.CustomerID
	if customerID == "" {
		if val, exists := ctx.Get("authenticated-user"); exists {
			customerID = val.(string)
		}
	}

	order, err := c.svc.CreateOrder(ctx.Request.Context(), customerID, dto.ProductID, dto.Quantity)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, order)
}

// GET /api/v1/orders/:id
func (c *OrderController) GetOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	order, err := c.svc.GetOrder(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, order)
}

// GET /api/v1/orders
func (c *OrderController) ListOrders(ctx *gin.Context) {
	orders, err := c.svc.ListOrders(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, orders)
}

// PUT /api/v1/orders/:id/status
func (c *OrderController) UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	var dto model.UpdateStatusDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := c.svc.UpdateOrderStatus(ctx.Request.Context(), id, dto.Status)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, order)
}
