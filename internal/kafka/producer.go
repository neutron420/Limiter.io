package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"limiter.io/internal/config"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	PublishEvent(ctx context.Context, key string, value []byte) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(cfg *config.Config) (Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.Hash{},
		MaxAttempts:  5,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne, // speed over durability, but structured
		Async:        true,             // Asynchronous writes
	}

	return &KafkaProducer{
		writer: writer,
	}, nil
}

func (kp *KafkaProducer) PublishEvent(ctx context.Context, key string, value []byte) error {
	err := kp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	return nil
}

func (kp *KafkaProducer) Close() error {
	if kp.writer != nil {
		if err := kp.writer.Close(); err != nil {
			log.Printf("Error closing Kafka writer: %v", err)
			return err
		}
	}
	return nil
}
