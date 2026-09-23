package out

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type CategoryProductsLoader interface {
	LoadByCategory(ctx context.Context, categoryID uuid.UUID) ([]*model.Product, error)
}
