package in

import (
	"context"
)

type OutboxPublisher interface {
	Execute(ctx context.Context, batchSize int) error
}
