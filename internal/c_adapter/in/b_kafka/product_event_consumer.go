package b_kafka

import (
	"context"
	"encoding/json"
	"log"

	"product-microservice/internal/a_domain/model"
	"product-microservice/internal/b_application/port/in"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ProductEventConsumer is the adapter for the input port for Kafka
type ProductEventConsumer struct {
	client  *kgo.Client
	useCase in.ProductProjector // The input port to synchronize the reading
}

// NewProductEventConsumer is the manual constructor
func NewProductEventConsumer(brokers []string, groupID string, topic string, useCase in.ProductProjector) (*ProductEventConsumer, error) {
	// Configure the franz-go client
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, err
	}

	return &ProductEventConsumer{
		client:  client,
		useCase: useCase,
	}, nil
}

// Start starts the infinite loop of consumption
func (c *ProductEventConsumer) ConsumeProductEvent(ctx context.Context) {
	log.Println("🚀 Kafka Consumer started and listening for events...")

	for {
		// 1. Fetch a batch of messages passing the context
		fetches := c.client.PollFetches(ctx)

		// If the main context is cancelled (e.g. server shutdown), we exit cleanly
		if ctx.Err() != nil {
			log.Println("🛑 Deteniendo el consumidor de Kafka...")
			return
		}

		// If there was any error in the network or connection with Kafka
		if errs := fetches.Errors(); len(errs) > 0 {
			log.Printf("⚠️ Error consuming from Kafka: %v", errs)
			continue
		}

		// 2. Process each message received
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			if record.Value == nil {
				c.projectDeletedProduct(ctx, record)
				continue
			}

			c.projectUpsertProduct(ctx, record)
		}
	}
}

// projectUpsertProduct applies a upsert: Key = product id, Value = product.
func (c *ProductEventConsumer) projectUpsertProduct(ctx context.Context, record *kgo.Record) {
	var product model.Product
	if err := json.Unmarshal(record.Value, &product); err != nil {
		log.Printf("❌ Error al deserializar mensaje de Kafka: %v", err)
		return
	}

	err := c.useCase.ProjectUpsert(ctx, &product)
	if err != nil {
		log.Printf("❌ Error processing product event in the service: %v", err)
		// Here you could implement retries or send to a Dead Letter Queue (DLQ)
		return
	}

	log.Printf("✅ Product upserted successfully for product ID: %s", product.ID)
}

// projectDeletedProduct applies a tombstone: Key = product id, Value = nil.
func (c *ProductEventConsumer) projectDeletedProduct(ctx context.Context, record *kgo.Record) {

	id, err := uuid.Parse(string(record.Key))
	if err != nil {
		log.Printf("❌ Invalid product id in Kafka key: %v", err)
		return
	}

	if err := c.useCase.ProjectDelete(ctx, id); err != nil {
		log.Printf("❌ Error processing product delete in the service: %v", err)
		return
	}

	log.Printf("✅ Producto eliminado de la proyección ID: %s", id)
}

// Close closes the connection to the Kafka client when the app is shut down
func (c *ProductEventConsumer) Close() {
	c.client.Close()
}
