package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DatabaseURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env", "../.env")

	host := getEnv("PRODUCT_DB_HOST", "localhost")
	port := getEnv("PRODUCT_DB_PORT", "5432")
	user := getEnv("PRODUCT_DB_USER", "postgres")
	pass := getEnv("PRODUCT_DB_PASSWORD", "Kumar578")
	name := getEnv("PRODUCT_DB_NAME", "products")
	ssl := getEnv("PRODUCT_DB_SSLMODE", "disable")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)

	return &Config{
		AppPort:     getEnv("PRODUCT_PORT", "8081"),
		DatabaseURL: dbURL,
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
