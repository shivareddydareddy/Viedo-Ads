// package kafka

// import (
//     "context"
//     "log"
//     "time"

//     "github.com/segmentio/kafka-go"
// )

// func PublishMessage(broker, topic string, message []byte) error {
//     writer := &kafka.Writer{
//         Addr:     kafka.TCP(broker),
//         Topic:    topic,
//         Balancer: &kafka.LeastBytes{},
//     }
//     defer writer.Close()

//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()

//     err := writer.WriteMessages(ctx, kafka.Message{
//         Value: message,
//     })
//     if err != nil {
//         log.Printf("Failed to write message to Kafka: %v", err)
//     }
//     return err
// }


package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

// InitKafkaWriter initializes the Kafka writer
func InitKafkaWriter(brokers []string, topic string) {
	writer = &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		RequiredAcks:           kafka.RequireOne,
		Async:                  true, // Non-blocking writes
		BatchSize:              100,
		BatchTimeout:           10 * time.Millisecond,
	}
}

// PublishMessage publishes a message to Kafka
func PublishMessage(broker, topic string, data []byte) error {
	// Create a new writer for each message if global writer is not initialized
	if writer == nil {
		w := &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			RequiredAcks: kafka.RequireOne,
		}
		defer w.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return w.WriteMessages(ctx, kafka.Message{
			Value: data,
		})
	}

	// Use global writer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

// CloseKafkaWriter closes the Kafka writer
func CloseKafkaWriter() {
	if writer != nil {
		if err := writer.Close(); err != nil {
			log.Printf("Error closing Kafka writer: %v", err)
		}
	}
}