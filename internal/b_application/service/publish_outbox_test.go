// Unit tests for the PublishOutbox relay (outbox → Kafka).
//
// The relay reads unpublished events, publishes to Kafka, then marks them published.
// We stub loader, marker, and publisher (hexagonal out ports) — no Kafka or Postgres
// needed. This is equivalent to @MockBean on repository + KafkaTemplate in Spring tests.
package service

import (
	"context"
	"encoding/json"
	"testing"

	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

// TestPublishOutbox_Execute_PublishesCreatedEvent covers product_created / product_updated:
// unmarshal payload → PublishProductUpsert → MarkAsPublished.
func TestPublishOutbox_Execute_PublishesCreatedEvent(t *testing.T) {
	productID := uuid.New()
	product := model.Product{
		ID:               productID,
		Name:             "Pasta dental",
		Price:            12.5,
		Stock:            10,
		Description:      "Pasta con fluor",
		ShortDescription: "Pasta",
		CategoryID:       uuid.New(),
	}
	payload, err := json.Marshal(product)
	if err != nil {
		t.Fatal(err)
	}

	eventID := uuid.New()
	loader := &stubOutboxLoader{events: []*model.OutboxEvent{{
		ID:          eventID,
		AggregateID: productID,
		EventType:   "product_created",
		Payload:     payload,
	}}}
	marker := &stubOutboxMarker{}
	publisher := &stubEventPublisher{}

	svc := NewPublishOutbox(loader, marker, publisher)
	if err := svc.Execute(context.Background(), 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if publisher.upserted == nil {
		t.Fatal("expected product upsert to be published")
	}
	if publisher.upserted.ID != productID {
		t.Fatalf("published product id = %s, want %s", publisher.upserted.ID, productID)
	}
	if len(marker.marked) != 1 || marker.marked[0] != eventID {
		t.Fatalf("expected event %s to be marked published", eventID)
	}
}

// TestPublishOutbox_Execute_PublishesDeleteEvent covers product_deleted:
// PublishProductDelete(tombstone) → MarkAsPublished.
func TestPublishOutbox_Execute_PublishesDeleteEvent(t *testing.T) {
	productID := uuid.New()
	eventID := uuid.New()
	loader := &stubOutboxLoader{events: []*model.OutboxEvent{{
		ID:          eventID,
		AggregateID: productID,
		EventType:   "product_deleted",
	}}}
	marker := &stubOutboxMarker{}
	publisher := &stubEventPublisher{}

	svc := NewPublishOutbox(loader, marker, publisher)
	if err := svc.Execute(context.Background(), 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if publisher.deleted != productID {
		t.Fatalf("deleted id = %s, want %s", publisher.deleted, productID)
	}
	if len(marker.marked) != 1 || marker.marked[0] != eventID {
		t.Fatalf("expected event %s to be marked published", eventID)
	}
}

type stubOutboxLoader struct {
	events []*model.OutboxEvent
	err    error
}

func (s *stubOutboxLoader) LoadUnpublished(_ context.Context, _ int) ([]*model.OutboxEvent, error) {
	return s.events, s.err
}

type stubOutboxMarker struct {
	marked []uuid.UUID
	err    error
}

func (s *stubOutboxMarker) MarkAsPublished(_ context.Context, id uuid.UUID) error {
	s.marked = append(s.marked, id)
	return s.err
}

type stubEventPublisher struct {
	upserted *model.Product
	deleted  uuid.UUID
}

func (s *stubEventPublisher) PublishProductUpsert(_ context.Context, product *model.Product) error {
	s.upserted = product
	return nil
}

func (s *stubEventPublisher) PublishProductDelete(_ context.Context, id uuid.UUID) error {
	s.deleted = id
	return nil
}
