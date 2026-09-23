package out

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"
)

type ProductSaver interface {
	Save(ctx context.Context, tx *sql.Tx, product *model.Product) error
}
