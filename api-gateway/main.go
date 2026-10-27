package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var (
	consulURL   string
	gatewayPort string
	jwtSecret   []byte
)

type LoginRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role" binding:"required"` // "ADMIN" or "CUSTOMER"
}

func init() {
	_ = godotenv.Load(".env", "../.env")
	consulURL = getEnv("CONSUL_HTTP_ADDR", "http://localhost:8500")
	gatewayPort = getEnv("GATEWAY_PORT", "8000")
	jwtSecret = []byte(getEnv("JWT_SECRET", "production-secret-key-change-me"))
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func resolveService(serviceName string, fallbackPort string) string {
	urlPath := fmt.Sprintf("%s/v1/health/service/%s?passing=true", consulURL, serviceName)
	resp, err := http.Get(urlPath)
	if err == nil {
		defer resp.Body.Close()
		var entries []struct {
			Service struct {
				Address string `json:"Address"`
				Port    int    `json:"Port"`
			} `json:"Service"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&entries); err == nil && len(entries) > 0 {
			addr := entries[0].Service.Address
			if addr == "" || addr == "127.0.0.1" {
				addr = "localhost"
			}
			return fmt.Sprintf("http://%s:%d", addr, entries[0].Service.Port)
		}
	}
	return "http://localhost:" + fallbackPort
}

func proxyTo(serviceName, fallbackPort string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetAddr := resolveService(serviceName, fallbackPort)
		targetURL, err := url.Parse(targetAddr)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Invalid service destination"})
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		c.Request.Host = targetURL.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP", "service": "api-gateway"})
	})

	r.POST("/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userId and role (ADMIN or CUSTOMER) are required"})
			return
		}

		claims := jwt.MapClaims{
			"user_id": req.UserID,
			"role":    req.Role,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
			"iat":     time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":  tokenString,
			"userId": req.UserID,
			"role":   req.Role,
		})
	})

	// GraphQL & Playground -> Order Service
	r.Any("/graphql", proxyTo("order-service", "8080"))
	r.Any("/playground", proxyTo("order-service", "8080"))

	// REST API Orders -> Order Service
	r.Any("/api/v1/orders/*path", proxyTo("order-service", "8080"))
	r.Any("/api/v1/orders", proxyTo("order-service", "8080"))

	// REST API Products -> Product Service
	r.Any("/products/*path", proxyTo("product-service", "8081"))
	r.Any("/products", proxyTo("product-service", "8081"))

	log.Printf("API Gateway active on port %s", gatewayPort)
	log.Fatal(r.Run(":" + gatewayPort))
}
