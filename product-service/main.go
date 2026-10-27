package main

import (
	"log"
	"net/http"
	"time"

	"product-service/config"
	"product-service/controller"
	"product-service/middleware"
	"product-service/repository"
	"product-service/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get DB object: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(15 * time.Minute)

	prodRepo, err := repository.NewProductRepository(db)
	if err != nil {
		log.Fatalf("Repo initialization failed: %v", err)
	}
	prodService := service.NewProductService(prodRepo)
	prodController := controller.NewProductController(prodService)

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.Recovery())

	// Health check for Consul
	r.GET("/health", func(c *gin.Context) {
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "UP", "service": "product-service"})
	})

	// Products routes:
	// 1. Browsing products requires valid authentication
	// 2. Creating products requires ADMIN authorization
	products := r.Group("/products")
	products.Use(middleware.RequireAuth())
	{
		products.GET("/:id", prodController.GetProduct)
		products.GET("", prodController.ListProducts)
		products.POST("", middleware.RequireRole("ADMIN"), prodController.CreateProduct)
	}

	log.Printf("Product Service running on port %s", cfg.AppPort)
	log.Fatal(r.Run(":" + cfg.AppPort))
}
