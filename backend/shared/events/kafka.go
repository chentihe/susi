package events

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(cfg KafkaConfig) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		Balancers: []kafka.Balancer{kafka.LeastBytes{}},
	})
	return &KafkaProducer{writer: writer}
}

func (p *KafkaProducer) Publish(event Event) error {
	msgBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: msgBytes,
	})
}

func (p *KafkaProducer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Println("Failed to close Kafka writer:", err)
	}
} 