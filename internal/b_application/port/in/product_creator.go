package in

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type ProductCreator interface {
	Execute(ctx context.Context, product *model.Product) (uuid.UUID, error)
}
