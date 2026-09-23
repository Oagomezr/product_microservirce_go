package out

import (
	"context"
	"product-microservice/internal/a_domain/model"
)

type ProductSearcher interface {
	Search(ctx context.Context, query string) ([]*model.Product, error)
}
