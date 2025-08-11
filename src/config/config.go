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
	MailerSend MailerSendConfig
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

type MailerSendConfig struct {
	Key           string
	SenderEmail   string
	SenderName    string
	OTPTemplateID string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT"),
			Mode: getEnv("SERVER_MODE"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET"),
			Expiry: getIntEnv("JWT_EXPIRY_MINUTE"),
		},
		RefreshJWT: RefreshJWTConfig{
			Secret: getEnv("REFRESH_JWT_SECRET"),
			Expiry: getIntEnv("REFRESH_JWT_EXPIRY_MINUTE"),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL"),
			File:  getEnv("LOG_FILE"),
		},
		S3: S3Config{
			Region:                   getEnv("S3_REGION"),
			Bucket:                   getEnv("S3_BUCKET_NAME"),
			PreSignedLifeTimeMinutes: getIntEnv("S3_PRESIGNED_LIFETIME_MINUTE"),
		},
		MailerSend: MailerSendConfig{
			Key:           getEnv("MAILER_KEY"),
			SenderEmail:   getEnv("MAILER_SENDER_EMAIL"),
			SenderName:    getEnv("MAILER_SENDER_NAME"),
			OTPTemplateID: getEnv("MAILER_OTP_TEMPLATE_ID"),
		},
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Fatalln("enviroment variable is missing: " + key)
	return ""
}

func getIntEnv(key string) int {
	if value, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return value
	}
	log.Fatalln("enviroment variable is missing: " + key)
	return 0
}
