package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port                       string
	Environment                string
	DatabaseURL                string
	FirebaseProjectID          string
	FirebaseStorageBucket      string
	FirebaseServiceAccountPath string
	FirebaseServiceAccountJSON string
	JWTSecret                  string
	MaxUploadSize              int64
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	maxUploadSize := int64(52428800) // 50MB default
	if size := os.Getenv("MAX_UPLOAD_SIZE"); size != "" {
		if parsed, err := strconv.ParseInt(size, 10, 64); err == nil {
			maxUploadSize = parsed
		}
	}

	return &Config{
		Port:                       getEnv("PORT", "8080"),
		Environment:                getEnv("ENV", "development"),
		DatabaseURL:                getEnv("DATABASE_PUBLIC_URL", ""),
		FirebaseProjectID:          getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseStorageBucket:      getEnv("FIREBASE_STORAGE_BUCKET", ""),
		FirebaseServiceAccountPath: getEnv("FIREBASE_SERVICE_ACCOUNT_PATH", ""),
		FirebaseServiceAccountJSON: getEnv("FIREBASE_SERVICE_ACCOUNT_JSON", ""),
		JWTSecret:                  getEnv("JWT_SECRET", ""),
		MaxUploadSize:              maxUploadSize,
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if c.FirebaseProjectID == "" {
		log.Fatal("FIREBASE_PROJECT_ID is required")
	}
	if c.FirebaseStorageBucket == "" {
		log.Fatal("FIREBASE_STORAGE_BUCKET is required")
	}
	if c.FirebaseServiceAccountPath == "" && c.FirebaseServiceAccountJSON == "" {
		log.Fatal("Either FIREBASE_SERVICE_ACCOUNT_PATH or FIREBASE_SERVICE_ACCOUNT_JSON is required")
	}
	return nil
}
