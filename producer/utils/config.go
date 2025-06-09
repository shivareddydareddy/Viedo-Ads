// package utils

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"
// )

// type Config struct {
//     KafkaBroker  string
//     KafkaTopic   string
//     RedisAddr    string
//     MongoURI     string
//     MongoDB      string
//     ProducerPort string
// }

// func LoadConfig() Config {
// 	fmt.Println("!!!!!!!!!!!@IN nev")
//     if err := godotenv.Load("/app/.env"); err != nil {
//         log.Println(err,"No .env file found, using environment variables")
//     }

//     return Config{
//         KafkaBroker:  getEnv("KAFKA_BROKER", "kafka:9092"),
//         KafkaTopic:   getEnv("KAFKA_TOPIC", "ad_clicks"),
//         RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
//         MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
//         MongoDB:      getEnv("MONGO_DB", "video_ads"),
//         ProducerPort: getEnv("PRODUCER_PORT", "8080"),
//     }
// }

// func getEnv(key, fallback string) string {
//     if value, exists := os.LookupEnv(key); exists {
//         return value
//     }
//     return fallback
// }


package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	KafkaBroker  string
	KafkaTopic   string
	RedisAddr    string
	MongoURI     string
	MongoDB      string
	ProducerPort string
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

	// Also try to find .env in current directory
	if workDir, err := os.Getwd(); err == nil {
		envFile := filepath.Join(workDir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			godotenv.Load(envFile)
		}
	}

	return Config{
		KafkaBroker:  getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "ad_clicks"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:      getEnv("MONGO_DB", "video_ads"),
		ProducerPort: getEnv("PRODUCER_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}