package out

import (
	"context"

	"github.com/google/uuid"
)

type OutboxMarker interface {
	MarkAsPublished(ctx context.Context, id uuid.UUID) error
}
