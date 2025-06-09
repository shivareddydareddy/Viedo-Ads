package kafka

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

func PublishMessage(broker, topic string, message []byte) error {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(broker),
        Topic:    topic,
        Balancer: &kafka.LeastBytes{},
    }
    defer writer.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := writer.WriteMessages(ctx, kafka.Message{
        Value: message,
    })
    if err != nil {
        log.Printf("Failed to write message to Kafka: %v", err)
    }
    return err
}