package out

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type ProductLoader interface {
	Load(ctx context.Context, id uuid.UUID) (*model.Product, error)
}
