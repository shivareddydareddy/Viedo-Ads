package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBroker   string
	KafkaTopic    string
	ConsumerGroup string
	MongoURI      string
	MongoDB       string
	RedisAddr     string
}

func LoadConfig() Config {
	// Try to load .env file from multiple locations
	envPaths := []string{
		".env",
		"/app/.env",
		"../.env",
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err == nil {
				log.Printf("Loaded environment from %s", path)
				break
			}
		}
	}
	if workDir, err := os.Getwd(); err == nil {
		envFile := filepath.Join(workDir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			godotenv.Load(envFile)
		}
	}

	return Config{
		KafkaBroker:   getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:    getEnv("KAFKA_TOPIC", "ad_clicks"),
		ConsumerGroup: getEnv("CONSUMER_GROUP", "ad_clicks_group"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017/"),
		MongoDB:       getEnv("MONGO_DB", "video_ads"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}