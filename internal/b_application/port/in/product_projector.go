package in

import (
	"context"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

type ProductProjector interface {
	ProjectUpsert(ctx context.Context, product *model.Product) error // Upsert is a combination of create and update
	ProjectDelete(ctx context.Context, id uuid.UUID) error           // Delete is a combination of soft delete and hard delete
}
