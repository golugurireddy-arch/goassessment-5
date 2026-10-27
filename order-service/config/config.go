package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort            string
	DatabaseURL        string
	ConsulAddr         string
	ProductServiceName string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env", "../.env")

	host := getEnv("ORDER_DB_HOST", "localhost")
	port := getEnv("ORDER_DB_PORT", "5432")
	user := getEnv("ORDER_DB_USER", "postgres")
	pass := getEnv("ORDER_DB_PASSWORD", "Kumar578")
	name := getEnv("ORDER_DB_NAME", "orders")
	ssl := getEnv("ORDER_DB_SSLMODE", "disable")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)

	return &Config{
		AppPort:            getEnv("ORDER_PORT", "8080"),
		DatabaseURL:        dbURL,
		ConsulAddr:         getEnv("CONSUL_HTTP_ADDR", "http://localhost:8500"),
		ProductServiceName: getEnv("PRODUCT_SERVICE_NAME", "product-service"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
