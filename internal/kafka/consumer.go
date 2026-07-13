package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/gorm/clause"

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
		MinBytes: 1,
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

	msgChan := make(chan kafka.Message)
	errChan := make(chan error)

	// Start a background worker to read messages blocking-ly without cancellations
	go func() {
		for {
			msg, err := kc.reader.ReadMessage(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errChan <- err:
				}
				continue
			}
			select {
			case <-ctx.Done():
				return
			case msgChan <- msg:
			}
		}
	}()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		log.Printf("Flushing %d analytics records to database...", len(buffer))
		err := kc.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&buffer).Error
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
		case msg := <-msgChan:
			var event models.AnalyticsLog
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Failed to unmarshal analytics log: %v", err)
				continue
			}

			buffer = append(buffer, event)
			if len(buffer) >= batchSize {
				flush()
			}
		case err := <-errChan:
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("Error reading message from Kafka: %v", err)
			time.Sleep(1 * time.Second)
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
