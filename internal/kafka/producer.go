package kafka

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
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

func CreateTopic(brokerAddr, topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	return controllerConn.CreateTopics(topicConfigs...)
}

func NewKafkaProducer(cfg *config.Config) (Producer, error) {
	// Programmatically create the topic if it does not exist
	if err := CreateTopic(cfg.KafkaBrokers, cfg.KafkaTopic); err != nil {
		log.Printf("Warning: Failed to auto-create topic %s: %v. It might already exist.", cfg.KafkaTopic, err)
	} else {
		log.Printf("Successfully verified/created Kafka topic: %s", cfg.KafkaTopic)
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers),
		Topic:        cfg.KafkaTopic,
		Balancer:     &kafka.Hash{},
		MaxAttempts:  5,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne, // speed over durability, but structured
		Async:        false,            // Synchronous write to catch errors in worker
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
