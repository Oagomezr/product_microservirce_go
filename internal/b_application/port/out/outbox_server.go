package out

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"
)

type OutboxSaver interface {
	Save(ctx context.Context, tx *sql.Tx, event *model.OutboxEvent) error
}
