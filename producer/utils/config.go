package utils

import (
	"fmt"
	"log"
	"os"

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
	fmt.Println("!!!!!!!!!!!@IN nev")
    if err := godotenv.Load("/app/.env"); err != nil {
        log.Println(err,"No .env file found, using environment variables")
    }

    return Config{
        KafkaBroker:  getEnv("KAFKA_BROKER", "kafka:9092"),
        KafkaTopic:   getEnv("KAFKA_TOPIC", "ad_clicks"),
        RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
        MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
        MongoDB:      getEnv("MONGO_DB", "video_ads"),
        ProducerPort: getEnv("PRODUCER_PORT", "8080"),
    }
}

func getEnv(key, fallback string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return fallback
}
