package kafka

import (
	"context"
	"fmt"
	"time"

	"encoding/json"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ProductEventPublisher is a struct that contains the producer for the product event publisher
type ProductEventPublisher struct {
	client *kgo.Client
	topic  string
}

// NewProductEventPublisher creates a new ProductEventPublisher
func NewProductEventPublisher(brokerURL []string, topic string) (*ProductEventPublisher, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokerURL...),
		kgo.DefaultProduceTopic(topic),
		// Fail produce attempts quickly when the broker is down (e.g. lab without Kafka).
		kgo.RecordDeliveryTimeout(8*time.Second),
		kgo.ProduceRequestTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}
	return &ProductEventPublisher{client: cl, topic: topic}, nil
}

// PublishProductCreate serializes the product create event and publishes it to the Kafka topic
func (p *ProductEventPublisher) PublishProductUpsert(ctx context.Context, product *model.Product) error {
	// 1. Become the model.Product to JSON
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product to JSON: %w", err)
	}

	// 2. Create the franz-go event record
	record := &kgo.Record{
		Key:       []byte(product.ID.String()),
		Value:     productJSON,
		Topic:     p.topic,
		Timestamp: time.Now(),
	}

	// 3. Produce the message to the synchronous way using the context and the record
	// franz-go manage the buffering and batching of the messages
	return p.client.ProduceSync(ctx, record).FirstErr()
}

func (p *ProductEventPublisher) PublishProductDelete(ctx context.Context, id uuid.UUID) error {
	// 1. Create the franz-go event record
	record := &kgo.Record{
		Key:       []byte(id.String()),
		Value:     nil,
		Topic:     p.topic,
		Timestamp: time.Now(),
	}

	// 2. Produce the message to the synchronous way using the context and the record
	// franz-go manage the buffering and batching of the messages
	return p.client.ProduceSync(ctx, record).FirstErr()
}

// Close closes the Kafka client
func (p *ProductEventPublisher) Close() {
	p.client.Close()
}
