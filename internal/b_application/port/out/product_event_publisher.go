package out

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type ProductEventPublisher interface {
	PublishProductUpsert(ctx context.Context, product *model.Product) error
	PublishProductDelete(ctx context.Context, id uuid.UUID) error
}
