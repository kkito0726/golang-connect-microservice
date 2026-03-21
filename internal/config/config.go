package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL       string
	Port              string
	ProductServiceURL string
	OrderServiceURL   string
	UserServiceURL    string
	PaymentServiceURL string
	JWTSecret         string
	JWTExpiryHours    int
}

func Load() (Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	jwtExpiryHours := 24
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("JWT_EXPIRY_HOURS must be an integer: %w", err)
		}
		jwtExpiryHours = n
	}

	return Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"),
		Port:              getEnv("PORT", "8080"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8080"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8083"),
		JWTSecret:         jwtSecret,
		JWTExpiryHours:    jwtExpiryHours,
	}, nil
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
