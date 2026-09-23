package outbox

import (
	"encoding/json"
	"product-microservice/internal/a_domain/model"
	"time"

	"github.com/google/uuid"
)

type OutboxEntity struct {
	ID          uuid.UUID
	AggregateID uuid.UUID
	EventType   string
	payload     json.RawMessage
	CreatedAt   time.Time
	Published   bool
}

func fromDomain(outboxEvent *model.OutboxEvent) *OutboxEntity {
	return &OutboxEntity{
		ID:          outboxEvent.ID,
		AggregateID: outboxEvent.AggregateID,
		EventType:   outboxEvent.EventType,
		payload:     outboxEvent.Payload,
		CreatedAt:   outboxEvent.CreatedAt,
		Published:   outboxEvent.Published,
	}
}

func toDomain(outboxEntity *OutboxEntity) *model.OutboxEvent {
	return &model.OutboxEvent{
		ID:          outboxEntity.ID,
		AggregateID: outboxEntity.AggregateID,
		EventType:   outboxEntity.EventType,
		Payload:     outboxEntity.payload,
		CreatedAt:   outboxEntity.CreatedAt,
		Published:   outboxEntity.Published,
	}
}
