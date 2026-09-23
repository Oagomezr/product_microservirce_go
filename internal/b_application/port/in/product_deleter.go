package in

import (
	"context"

	"github.com/google/uuid"
)

type ProductDeleter interface {
	Execute(ctx context.Context, id uuid.UUID) error
}
