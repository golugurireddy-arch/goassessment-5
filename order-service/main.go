package main

import (
	"log"
	"net/http"
	"order-service/client"
	"order-service/config"
	"order-service/controller"
	"order-service/graph"
	"order-service/middleware"
	"order-service/repository"
	"order-service/service"

	"github.com/gin-gonic/gin"
	gqlhandler "github.com/graphql-go/handler"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	orderRepo, err := repository.NewOrderRepository(db)
	if err != nil {
		log.Fatalf("Repository error: %v", err)
	}

	productClient := client.NewResilientProductClient(cfg.ConsulAddr, cfg.ProductServiceName)
	orderService := service.NewOrderService(orderRepo, productClient)
	orderController := controller.NewOrderController(orderService)

	gqlSchema, err := graph.BuildSchema(orderService)
	if err != nil {
		log.Fatalf("GraphQL init error: %v", err)
	}

	h := gqlhandler.New(&gqlhandler.Config{
		Schema:     &gqlSchema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.Recovery())

	// Consul health check
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN", "error": "database unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "UP", "service": "order-service"})
	})

	// GraphQL endpoints (No authentication required)
	r.Any("/playground", gin.WrapH(h))
	r.Any("/graphql", gin.WrapH(h))

	// REST API v1 endpoints
	v1 := r.Group("/api/v1/orders")
	v1.Use(middleware.RequireAuth())
	v1.POST("", orderController.CreateOrder)
	v1.GET("/:id", middleware.RequireOrderAccess(orderRepo), orderController.GetOrder)
	v1.GET("", middleware.RequireRole("ADMIN"), orderController.ListOrders)
	v1.PUT("/:id/status", middleware.RequireRole("ADMIN"), orderController.UpdateStatus)

	log.Printf("Order Service active on port %s (REST + GraphQL)", cfg.AppPort)
	log.Fatal(r.Run(":" + cfg.AppPort))
}
