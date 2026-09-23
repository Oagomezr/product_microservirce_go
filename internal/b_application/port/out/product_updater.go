package out

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"
)

type ProductUpdater interface {
	Update(ctx context.Context, tx *sql.Tx, product *model.Product) error
}
