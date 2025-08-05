package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database   DatabaseConfig
	Server     ServerConfig
	JWT        JWTConfig
	RefreshJWT RefreshJWTConfig
	Logging    LoggingConfig
	S3         S3Config
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Mode string
}

type JWTConfig struct {
	Secret string
	Expiry int
}

type RefreshJWTConfig struct {
	Secret string
	Expiry int
}

type LoggingConfig struct {
	Level string
	File  string
}

type S3Config struct {
	Region                   string
	Bucket                   string
	PreSignedLifeTimeMinutes int
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "sama_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-here"),
			Expiry: getIntEnv("JWT_EXPIRY", 7*24*60), // 1 day
		},
		RefreshJWT: RefreshJWTConfig{
			Secret: getEnv("REFRESH_JWT_SECRET", "your-secret-key-here"),
			Expiry: getIntEnv("REFRESH_JWT_EXPIRY", 30*24*60), // 1 month
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", "logs/app.log"),
		},
		S3: S3Config{
			Region:                   getEnv("S3_REGION", "ap-northeast-2"),
			Bucket:                   getEnv("S3_BUCKET_NAME", "test-bucket"),
			PreSignedLifeTimeMinutes: getIntEnv("S3_PRESIGNED_LIFETIME_MINUTE", 15),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	return defaultValue
}
