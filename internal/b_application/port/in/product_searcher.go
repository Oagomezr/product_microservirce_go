package in

import (
	"context"
	"product-microservice/internal/a_domain/model"
)

type ProductSearcher interface {
	Execute(ctx context.Context, query string) ([]*model.Product, error)
}
