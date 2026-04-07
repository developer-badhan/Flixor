package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

/**
 * Config holds all environment variables for the application.
 * Every other package import this, never call os.Getenv() directly elsewhere.
 */
type Config struct {
	Port           string
	AppEnv         string
	MongoURI       string
	DBName         string
	JWTSecret      string
	JWTExpiryHours int
	GEMINI_API_KEY string
}

/**
 * Load reads the .env file and returns a populated Config struct.
 * Call this once in main.go at startup - pass the result everywhere.
 */
func Load() *Config {
    // Try to load .env
    if err := godotenv.Load(".env"); err != nil {
        log.Fatalf("❌ Failed to load .env file: %v", err)
    } else {
        log.Println("✅ Successfully loaded .env file")
    }

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		log.Fatal("JWT_EXPIRY_HOURS must be a valid integer")
	}

	cfg := &Config{
		Port:           getEnv("PORT", "5000"),
		AppEnv:         getEnv("APP_ENV", "development"),
		MongoURI:       getEnv("MONGO_URI", ""),
		DBName:         getEnv("DB_NAME", "flixor"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: jwtExpiry,
		GEMINI_API_KEY: getEnv("GEMINI_API_KEY", ""),
	}

	// Validate required fields, crash early with a clear message
	cfg.validate()

	return cfg
}

/**
 * validate() - checks that all required config values are present.
 * The app should not start if critical config is missing.
 */
func (c *Config) validate() {
	if c.MongoURI == "" {
		log.Fatal("MONGO_URI is required but not set in environment")
	}
	if c.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required but not set in environment")
	}
	if c.JWTSecret == "change_this_to_a_long_random_secret" {
		log.Println("WARNING: JWT_SECRET is using the default placeholder — change it before production")
	}
	if c.GEMINI_API_KEY == "" {
		log.Fatal("GEMINI_API_KEY is required but not set in environment")
	}
}

/**
 * getEnv() - reads an environemnt variable, falling back to a default value
 * If it is not set, keeps every filed above clean, no inline if-checks.
 */
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
