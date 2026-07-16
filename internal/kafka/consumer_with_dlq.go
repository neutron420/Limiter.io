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

type KafkaConsumerWithDLQ struct {
	reader         *kafka.Reader
	db             *gorm.DB
	dlq            *DLQProducer
	dlqTopic       string
	maxProcessingRetries int
}

func NewKafkaConsumerWithDLQ(cfg *config.Config, db *gorm.DB) Consumer {
	dlqTopic := cfg.KafkaTopic + "_dlq"
	dlq := NewDLQProducer(cfg.KafkaBrokers, dlqTopic)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.KafkaBrokers},
		Topic:    cfg.KafkaTopic,
		GroupID:  cfg.KafkaGroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})

	return &KafkaConsumerWithDLQ{
		reader:         reader,
		db:             db,
		dlq:            dlq,
		dlqTopic:       dlqTopic,
		maxProcessingRetries: 3,
	}
}

func (kc *KafkaConsumerWithDLQ) Start(ctx context.Context) error {
	log.Println("Starting Kafka consumer with DLQ loop...")
	const batchSize = 100
	const flushInterval = 2 * time.Second

	buffer := make([]models.AnalyticsLog, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	msgChan := make(chan kafka.Message)
	errChan := make(chan error)

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
			for _, record := range buffer {
				payload, _ := json.Marshal(record)
				_ = kc.dlq.SendToDLQ(context.Background(), kc.reader.Config().Topic, record.ID.String(), payload, err.Error(), 0)
			}
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
				log.Printf("Failed to unmarshal analytics log, sending to DLQ: %v", err)
				_ = kc.dlq.SendToDLQ(ctx, kc.reader.Config().Topic, string(msg.Key), msg.Value, err.Error(), 0)
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

func (kc *KafkaConsumerWithDLQ) Close() error {
	if kc.reader != nil {
		_ = kc.reader.Close()
	}
	if kc.dlq != nil {
		_ = kc.dlq.Close()
	}
	return nil
}
