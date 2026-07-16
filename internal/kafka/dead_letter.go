package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type DeadLetterMessage struct {
	OriginalTopic string          `json:"original_topic"`
	OriginalKey   string          `json:"original_key"`
	OriginalValue json.RawMessage `json:"original_value"`
	Error         string          `json:"error"`
	FailedAt      time.Time       `json:"failed_at"`
	RetryCount    int             `json:"retry_count"`
}

type DLQProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewDLQProducer(brokers, dlqTopic string) *DLQProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Topic:        dlqTopic,
		Balancer:     &kafka.Hash{},
		MaxAttempts:  3,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
	if err := CreateTopic(brokers, dlqTopic); err != nil {
		log.Printf("Warning: DLQ topic %s may already exist: %v", dlqTopic, err)
	}
	return &DLQProducer{writer: writer, topic: dlqTopic}
}

func (d *DLQProducer) SendToDLQ(ctx context.Context, originalTopic string, originalKey string, originalValue []byte, errMsg string, retryCount int) error {
	dlqMsg := DeadLetterMessage{
		OriginalTopic: originalTopic,
		OriginalKey:   originalKey,
		OriginalValue: originalValue,
		Error:         errMsg,
		FailedAt:      time.Now().UTC(),
		RetryCount:    retryCount,
	}
	payload, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}
	return d.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(originalKey),
		Value: payload,
	})
}

func (d *DLQProducer) Close() error {
	if d.writer != nil {
		return d.writer.Close()
	}
	return nil
}

func RetryWithDLQ(producer *KafkaProducer, dlq *DLQProducer, ctx context.Context, key string, value []byte, maxRetries int) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			backoff := time.Duration(1<<i) * time.Second
			log.Printf("Retry %d/%d for key %s after %v", i+1, maxRetries, key, backoff)
			time.Sleep(backoff)
		}
		lastErr = producer.PublishEvent(ctx, key, value)
		if lastErr == nil {
			return
		}
		log.Printf("Failed to publish event (attempt %d/%d): %v", i+1, maxRetries, lastErr)
	}
	if dlq != nil {
		dlqErr := dlq.SendToDLQ(ctx, producer.writer.Topic, key, value, lastErr.Error(), maxRetries)
		if dlqErr != nil {
			log.Printf("Failed to send to DLQ: %v", dlqErr)
		} else {
			log.Printf("Sent message to DLQ: key=%s", key)
		}
	}
}
