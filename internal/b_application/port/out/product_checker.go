package out

import (
	"context"

	"github.com/google/uuid"
)

type ProductChecker interface {
	Exist(ctx context.Context, id uuid.UUID) (bool, error)
}
