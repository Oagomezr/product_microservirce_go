package out

import (
	"context"

	"github.com/google/uuid"
)

type ProductProjectionDeleter interface {
	Delete(ctx context.Context, id uuid.UUID) error
}
