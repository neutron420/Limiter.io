package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"limiter.io/internal/config"
	"limiter.io/internal/models"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type KafkaConsumer struct {
	reader *kafka.Reader
	db     *gorm.DB
}

func NewKafkaConsumer(cfg *config.Config, db *gorm.DB) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBrokers},
		Topic:    cfg.KafkaTopic,
		GroupID:  cfg.KafkaGroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		MaxWait:  1 * time.Second,
	})

	return &KafkaConsumer{
		reader: reader,
		db:     db,
	}
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	log.Println("Starting Kafka consumer loop...")
	const batchSize = 100
	const flushInterval = 2 * time.Second

	buffer := make([]models.AnalyticsLog, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		log.Printf("Flushing %d analytics records to database...", len(buffer))
		err := kc.db.Create(&buffer).Error
		if err != nil {
			log.Printf("Error batch inserting analytics records: %v", err)
		} else {
			log.Printf("Successfully inserted %d analytics records", len(buffer))
		}
		buffer = buffer[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return ctx.Err()
		case <-ticker.C:
			flush()
		default:
			// Read message with a timeout context to allow tick / cancellation check
			readCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			msg, err := kc.reader.ReadMessage(readCtx)
			cancel()

			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				log.Printf("Error reading message from Kafka: %v", err)
				time.Sleep(1 * time.Second) // pause briefly on connection errors
				continue
			}

			var event models.AnalyticsLog
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Failed to unmarshal analytics log: %v", err)
				continue
			}

			buffer = append(buffer, event)
			if len(buffer) >= batchSize {
				flush()
			}
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	if kc.reader != nil {
		if err := kc.reader.Close(); err != nil {
			log.Printf("Error closing Kafka reader: %v", err)
			return err
		}
	}
	return nil
}
