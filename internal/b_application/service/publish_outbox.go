package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"product-microservice/internal/a_domain/model"
	"product-microservice/internal/b_application/port/out"
)

type PublishOutbox struct {
	loader    out.OutboxLoader
	marker    out.OutboxMarker
	publisher out.ProductEventPublisher
}

func NewPublishOutbox(loader out.OutboxLoader, marker out.OutboxMarker, publisher out.ProductEventPublisher) *PublishOutbox {
	return &PublishOutbox{
		loader:    loader,
		marker:    marker,
		publisher: publisher,
	}
}

func (s *PublishOutbox) Execute(ctx context.Context, batchSize int) error {
	batch, err := s.loader.LoadUnpublished(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("load unpublished outbox: %w", err)
	}

	log.Printf("outbox relay: tick, unpublished=%d", len(batch))
	if len(batch) == 0 {
		return nil
	}

	for _, message := range batch {
		switch message.EventType {
		case "product_created", "product_updated":
			var product model.Product
			err := json.Unmarshal(message.Payload, &product)
			if err != nil {
				log.Printf("outbox relay: unmarshal failed event_id=%s: %v", message.ID, err)
				continue
			}
			err = s.publisher.PublishProductUpsert(ctx, &product)
			if err != nil {
				log.Printf("outbox relay: kafka publish failed event_id=%s type=%s: %v", message.ID, message.EventType, err)
				continue
			}
			err = s.marker.MarkAsPublished(ctx, message.ID)
			if err != nil {
				log.Printf("outbox relay: mark published failed event_id=%s: %v", message.ID, err)
				continue
			}
			log.Printf("outbox relay: published ok event_id=%s type=%s product_id=%s", message.ID, message.EventType, product.ID)

		case "product_deleted":
			err = s.publisher.PublishProductDelete(ctx, message.AggregateID)
			if err != nil {
				log.Printf("outbox relay: kafka delete publish failed event_id=%s: %v", message.ID, err)
				continue
			}
			err = s.marker.MarkAsPublished(ctx, message.ID)
			if err != nil {
				log.Printf("outbox relay: mark published failed event_id=%s: %v", message.ID, err)
				continue
			}
			log.Printf("outbox relay: published ok event_id=%s type=%s product_id=%s", message.ID, message.EventType, message.AggregateID)

		default:
			log.Printf("outbox relay: invalid event type=%s event_id=%s", message.EventType, message.ID)
			continue
		}
	}
	return nil
}
