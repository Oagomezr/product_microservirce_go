package in

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type ProductFinder interface {
	Execute(ctx context.Context, id uuid.UUID) (*model.Product, error)
}
