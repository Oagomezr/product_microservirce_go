package out

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ProductDeleter interface {
	Delete(ctx context.Context, tx *sql.Tx, id uuid.UUID) error
}
