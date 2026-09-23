package out

import (
	"context"
	"product-microservice/internal/a_domain/model"
)

type OutboxLoader interface {
	LoadUnpublished(ctx context.Context, limit int) ([]*model.OutboxEvent, error)
}
