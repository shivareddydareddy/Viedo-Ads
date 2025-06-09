package kafka

import (
	"consumer/db"
	"consumer/services"
	"consumer/utils"
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// Remove the local ClickEvent definition and use services.ClickEvent

type Consumer struct {
	reader      *kafka.Reader
	mongoClient *db.MongoClient
	redisClient *db.RedisClient
	wg          sync.WaitGroup
	config      utils.Config
}

func StartConsumer(ctx context.Context, cfg utils.Config, mongoClient *db.MongoClient, redisClient *db.RedisClient) error {
	// Create Kafka reader with improved configuration
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(cfg.KafkaBroker, ","), // Support multiple brokers
		GroupID: cfg.ConsumerGroup,
		Topic:   cfg.KafkaTopic,
		
		// Performance tuning
		MinBytes:    1,        // Minimum bytes to read
		MaxBytes:    10e6,     // Maximum bytes to read (10MB)
		MaxWait:     500 * time.Millisecond, // Max wait time for batch
		
		// Reliability settings
		CommitInterval: 1 * time.Second,     // Auto-commit interval
		StartOffset:    kafka.LastOffset,    // Start from latest messages
		
		// Error handling
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("Kafka Reader Error: "+msg, args...)
		}),
		
		// Partition management
		Partition: 0, // Can be removed to consume from all partitions
	})

	consumer := &Consumer{
		reader:      reader,
		mongoClient: mongoClient,
		redisClient: redisClient,
		config:      cfg,
	}

	defer func() {
		log.Println("Closing Kafka reader...")
		if err := reader.Close(); err != nil {
			log.Printf("Error closing Kafka reader: %v", err)
		}
		consumer.wg.Wait() // Wait for all goroutines to complete
		log.Println("All processing goroutines completed")
	}()

	log.Printf("Kafka consumer started. Brokers: %s, Topic: %s, Group: %s", 
		cfg.KafkaBroker, cfg.KafkaTopic, cfg.ConsumerGroup)

	// Message processing loop
	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer...")
			return ctx.Err()
			
		default:
			// Set a timeout for reading messages
			msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			
			message, err := reader.ReadMessage(msgCtx)
			cancel()
			
			if err != nil {
				if err == context.DeadlineExceeded {
					// This is normal when no messages are available
					continue
				}
				log.Printf("Error reading Kafka message: %v", err)
				
				// Exponential backoff on errors
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
					continue
				}
			}

			messageCount++
			log.Printf("Received message #%d from partition %d, offset %d", 
				messageCount, message.Partition, message.Offset)

			// Process message in goroutine for better throughput
			consumer.wg.Add(1)
			go consumer.processMessage(message)
		}
	}
}

func (c *Consumer) processMessage(message kafka.Message) {
	defer c.wg.Done()

	// Use services.ClickEvent directly
	var event services.ClickEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		log.Printf("Failed to unmarshal click event: %v, raw message: %s", err, string(message.Value))
		return
	}

	log.Printf("Processing click event: AdID=%s, IP=%s, PlaybackSeconds=%d", 
		event.AdID, event.IP, event.PlaybackSeconds)

	// Process the click event - now types match!
	if err := services.ProcessClickEvent(event, c.mongoClient, c.redisClient); err != nil {
		log.Printf("Failed to process click event: %v", err)
		// In production, you might want to send to a dead letter queue
		return
	}

	log.Printf("Successfully processed click event for AdID: %s", event.AdID)
}