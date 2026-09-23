package in

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type CategoryProductLister interface {
	Execute(ctx context.Context, categoryID uuid.UUID) ([]*model.Product, error)
}
