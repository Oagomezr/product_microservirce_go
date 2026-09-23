package out

import (
	"context"
	"product-microservice/internal/a_domain/model"
)

type ProductProjectionSaver interface {
	Upsert(ctx context.Context, product *model.Product) error
}
