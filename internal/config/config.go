package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL       string
	Port              string
	ProductServiceURL string
	OrderServiceURL   string
	UserServiceURL    string
	PaymentServiceURL string
}

func Load() Config {
	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"),
		Port:              getEnv("PORT", "8080"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8080"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083"),
	}
}

func (c Config) DatabaseURLWithName(dbName string) string {
	return fmt.Sprintf("postgres://postgres:postgres@localhost:5432/%s?sslmode=disable", dbName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
