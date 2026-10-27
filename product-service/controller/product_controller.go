package controller

import (
	"errors"
	"net/http"
	"strconv"

	"product-service/model"
	"product-service/repository"
	"product-service/service"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	svc service.ProductService
}

func NewProductController(svc service.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

func (ctrl *ProductController) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID must be an integer"})
		return
	}

	prod, err := ctrl.svc.GetProduct(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prod)
}

func (ctrl *ProductController) ListProducts(c *gin.Context) {
	products, err := ctrl.svc.ListProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prod, err := ctrl.svc.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, prod)
}
