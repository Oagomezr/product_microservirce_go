package service

import (
	"context"
	"encoding/json"
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
		return err
	}
	for _, message := range batch {

		switch message.EventType {
		case "product_created", "product_updated":
			var product model.Product
			err := json.Unmarshal(message.Payload, &product)
			if err != nil {
				log.Printf("failed to unmarshal product: %v", err)
				continue
			}
			err = s.publisher.PublishProductUpsert(ctx, &product)
			if err != nil {
				log.Printf("failed to publish product upsert: %v", err)
				continue
			}
			err = s.marker.MarkAsPublished(ctx, message.ID)
			if err != nil {
				log.Printf("failed to mark as published: %v", err)
				continue
			}

		case "product_deleted":
			err = s.publisher.PublishProductDelete(ctx, message.AggregateID)
			if err != nil {
				log.Printf("failed to publish product delete: %v", err)
				continue
			}
			err = s.marker.MarkAsPublished(ctx, message.ID)
			if err != nil {
				log.Printf("failed to mark as published: %v", err)
				continue
			}

		default:
			log.Printf("invalid event type: %s", message.EventType)
			continue

		}

	}
	return nil
}
