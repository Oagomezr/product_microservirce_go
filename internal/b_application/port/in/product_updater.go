package in

import (
	"context"
	"product-microservice/internal/a_domain/model"
)

type ProductUpdater interface {
	Execute(ctx context.Context, product *model.ProductPatch) error
}
