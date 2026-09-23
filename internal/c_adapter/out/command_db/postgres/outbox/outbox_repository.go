package outbox

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, tx *sql.Tx, outboxEvent *model.OutboxEvent) error {
	outboxEntity := fromDomain(outboxEvent)
	_, err := tx.ExecContext(ctx, "INSERT INTO outbox (id, aggregate_id, event_type, payload, created_at, published) VALUES ($1, $2, $3, $4, $5, $6)", outboxEntity.ID, outboxEntity.AggregateID, outboxEntity.EventType, outboxEntity.payload, outboxEntity.CreatedAt, outboxEntity.Published)
	return err
}

func (r *OutboxRepository) LoadUnpublished(ctx context.Context, limit int) ([]*model.OutboxEvent, error) {

	rows, err := r.db.QueryContext(ctx, "SELECT id, aggregate_id, event_type, payload, created_at, published FROM outbox WHERE published = false ORDER BY created_at ASC LIMIT $1", limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outboxEvents []*model.OutboxEvent
	for rows.Next() {
		var outboxEntity OutboxEntity
		err := rows.Scan(&outboxEntity.ID, &outboxEntity.AggregateID, &outboxEntity.EventType, &outboxEntity.payload, &outboxEntity.CreatedAt, &outboxEntity.Published)
		if err != nil {
			return nil, err
		}
		outboxEvents = append(outboxEvents, toDomain(&outboxEntity))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return outboxEvents, nil
}

func (r *OutboxRepository) MarkAsPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE outbox SET published = true WHERE id = $1", id)
	return err
}
