// package main

// import (
// 	"consumer/kafka"
// 	"consumer/utils"
// 	"log"
// 	// "video-ads-backend/consumer/kafka"
// 	// "video-ads-backend/consumer/utils"
// )

// func main() {
//     cfg := utils.LoadConfig()
//     log.Println("Starting Consumer Service...")
//     kafka.StartConsumer(cfg)
// }


package main

import (
	"consumer/db"
	"consumer/kafka"
	"consumer/utils"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting Consumer Service...")
	
	// Load configuration
	cfg := utils.LoadConfig()
	log.Printf("Config loaded: Kafka=%s, Topic=%s, Group=%s", 
		cfg.KafkaBroker, cfg.KafkaTopic, cfg.ConsumerGroup)

	// Initialize database connections with retries
	mongoClient, err := initializeMongoWithRetry(cfg, 5)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB after retries: %v", err)
	}
	defer mongoClient.Disconnect()
	log.Println("Connected to MongoDB successfully")

	redisClient, err := initializeRedisWithRetry(cfg, 5)
	if err != nil {
		log.Fatalf("Failed to connect to Redis after retries: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis successfully")

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping consumer...")
		cancel()
	}()

	// Start Kafka consumer
	log.Println("Starting Kafka consumer...")
	if err := kafka.StartConsumer(ctx, cfg, mongoClient, redisClient); err != nil {
		log.Fatalf("Consumer failed: %v", err)
	}

	log.Println("Consumer service stopped")
}

func initializeMongoWithRetry(cfg utils.Config, maxRetries int) (*db.MongoClient, error) {
	var mongoClient *db.MongoClient
	var err error

	for i := 0; i < maxRetries; i++ {
		mongoClient, err = db.NewMongoClient(cfg.MongoURI, cfg.MongoDB)
		if err == nil {
			return mongoClient, nil
		}
		
		log.Printf("MongoDB connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
		}
	}
	
	return nil, err
}

func initializeRedisWithRetry(cfg utils.Config, maxRetries int) (*db.RedisClient, error) {
	var redisClient *db.RedisClient
	var err error

	for i := 0; i < maxRetries; i++ {
		redisClient, err = db.NewRedisClient(cfg.RedisAddr)
		if err == nil {
			return redisClient, nil
		}
		
		log.Printf("Redis connection attempt %d failed: %v", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
		}
	}
	
	return nil, err
}