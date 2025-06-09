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



type Consumer struct {
	reader      *kafka.Reader
	mongoClient *db.MongoClient
	redisClient *db.RedisClient
	wg          sync.WaitGroup
	config      utils.Config
}

func StartConsumer(ctx context.Context, cfg utils.Config, mongoClient *db.MongoClient, redisClient *db.RedisClient) error {
	
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(cfg.KafkaBroker, ","),
		GroupID: cfg.ConsumerGroup,
		Topic:   cfg.KafkaTopic,
		
		MinBytes:    1,        
		MaxBytes:    10e6,     
		MaxWait:     500 * time.Millisecond, 
		CommitInterval: 1 * time.Second,   
		StartOffset:    kafka.LastOffset,   
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("Kafka Reader Error: "+msg, args...)
		}),

		Partition: 0, 
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
			log.Printf("Error closing reader: %v", err)
		}
		consumer.wg.Wait() 
		log.Println("All processing goroutines completed")
	}()

	log.Printf("Kafka consumer started. Brokers: %s, Topic: %s, Group: %s", 
		cfg.KafkaBroker, cfg.KafkaTopic, cfg.ConsumerGroup)

	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer...")
			return ctx.Err()
			
		default:
			
			msgCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			
			message, err := reader.ReadMessage(msgCtx)
			cancel()
			
			if err != nil {
				if err == context.DeadlineExceeded {
					
					continue
				}
				log.Printf("Error reading Kafka message: %v", err)
				
				
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

			
			consumer.wg.Add(1)
			go consumer.processMessage(message)
		}
	}
}

func (c *Consumer) processMessage(message kafka.Message) {
	defer c.wg.Done()

	var event services.ClickEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		log.Printf("Failed to unmarshal click event: %v, raw message: %s", err, string(message.Value))
		return
	}

	log.Printf("Processing click event: AdID=%s, IP=%s, PlaybackSeconds=%d", 
		event.AdID, event.IP, event.PlaybackSeconds)

	if err := services.ProcessClickEvent(event, c.mongoClient, c.redisClient); err != nil {
		log.Printf("Failed to process click event: %v", err)		
		return
	}

	log.Printf("Successfully processed click event for AdID: %s", event.AdID)
}