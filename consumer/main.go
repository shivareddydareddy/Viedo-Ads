package main

import (
	"consumer/db"
	"consumer/kafka"
	"consumer/services"
	"consumer/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"

	// "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	
	log.Println("Starting Consumer Service...")
	
	cfg := utils.LoadConfig()
	log.Printf("Config loaded: Kafka=%s, Topic=%s, Group=%s", 
		cfg.KafkaBroker, cfg.KafkaTopic, cfg.ConsumerGroup)

	fmt.Println("congifff",cfg)
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

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping consumer...")
		defer cancel()
	}()

    go func() {
        err := kafka.StartConsumer(ctx, cfg, mongoClient, redisClient)
        if err != nil {
            log.Fatalf("Kafka consumer failed: %v", err)
        }
    }()	

	fmt.Println("@@@@@@@@@@@@@@@@@@@@@@")
	startHTTPServer(ctx, mongoClient, redisClient)

	// log.Println("Consumer service stopped")
}

func startHTTPServer(ctx context.Context, mongoClient *db.MongoClient, redisClient *db.RedisClient) {
	app := fiber.New()
// 	mongoClient := &services.MongoClient{
// 	Client:   actualMongoClient,
// 	Database: actualMongoDatabase,
// }

	app.Get("/ads", func(c *fiber.Ctx) error {
		const redisKey = "ads:all"
		fmt.Println("@@@@@@@@@@@@@@@@@@")

		val, err := redisClient.Get(ctx, redisKey)
		if err == nil && val != "" {
			var ads []map[string]interface{}
			if err := json.Unmarshal([]byte(val), &ads); err == nil {
				return c.JSON(ads)
			}
		}

		ads, err := mongoClient.GetAllAds()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ads from DB"})
		}

		data, _ := json.Marshal(ads)
		redisClient.Set(ctx, redisKey, data, time.Hour)
		return c.JSON(ads)
	})

	app.Get("/ads/analytics", func(c *fiber.Ctx) error {
		adID := c.Query("id")
		window := c.Query("window", "1h")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")

		dur, err := time.ParseDuration(window)
		fmt.Println("timeeeee",dur)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid window"})
		}

		end := time.Now()
		start := end.Add(-dur)
		analytics, err := services.GetAdAnalytics(mongoClient,adID, end.Sub(start), redisClient)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch analytics"})
		}

		return c.JSON(analytics)
	})

	log.Println("Fiber HTTP server running on :8081")
	if err := app.Listen(":8082"); err != nil {
		log.Fatalf("Fiber server failed: %v", err)
	}
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
