package product

import (
	"context"
	"database/sql"
	"fmt"
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

// 1. The only one adapter that implements the ProductRepository interface
type ProductRepository struct {
	db *sql.DB
}

// 2. Constructor adapter
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// 3. Equivalent to my "DeleteProductAdapter.java"
// When writing this method, this struct already implements automatically the deletePort(product_deleter) interface
func (r *ProductRepository) Delete(ctx context.Context, tx *sql.Tx, productID uuid.UUID) error {
	_, err := tx.ExecContext(
		ctx,
		"DELETE FROM products WHERE id = $1",
		productID,
	)
	return err
}

// 4. Equivalent to my "UpdateProductAdapter.java"
// When writing this method, this struct already implements automatically the updatePort(product_updater) interface
func (r *ProductRepository) Update(ctx context.Context, tx *sql.Tx, product *model.Product) error {
	productEntity := fromDomain(product)

	query := `
		UPDATE products
		SET name = $1, price = $2, stock = $3, description = $4, short_description = $5, hidden = $6, updated_at = $7, images_urls = $8, category_id = $9
		WHERE id = $10
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		productEntity.Name,
		productEntity.Price,
		productEntity.Stock,
		productEntity.Description,
		productEntity.ShortDescription,
		productEntity.Hidden,
		productEntity.UpdatedAt,
		productEntity.ImagesURLs,
		productEntity.CategoryID,
		productEntity.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

// 5. Equivalent to my "CreateProductAdapter.java"
// When writing this method, this struct already implements automatically the createPort(product_creator) interface
func (r *ProductRepository) Save(ctx context.Context, tx *sql.Tx, product *model.Product) error {
	productEntity := fromDomain(product)

	query := `
		INSERT INTO products (id, name, price, stock, description, short_description, hidden, created_at, updated_at, images_urls, category_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		productEntity.ID,
		productEntity.Name,
		productEntity.Price,
		productEntity.Stock,
		productEntity.Description,
		productEntity.ShortDescription,
		productEntity.Hidden,
		productEntity.CreatedAt,
		productEntity.UpdatedAt,
		productEntity.ImagesURLs,
		productEntity.CategoryID,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE id = $1", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if product exists: %w", err)
	}
	return count > 0, nil
}

func (r *ProductRepository) Load(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	var productEntity ProductEntity
	err := r.db.QueryRowContext(ctx, "SELECT * FROM products WHERE id = $1", id).Scan(
		&productEntity.ID,
		&productEntity.Name,
		&productEntity.Price,
		&productEntity.Stock,
		&productEntity.Description,
		&productEntity.ShortDescription,
		&productEntity.Hidden,
		&productEntity.CreatedAt,
		&productEntity.UpdatedAt,
		&productEntity.ImagesURLs,
		&productEntity.CategoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load product: %w", err)
	}
	return toDomain(&productEntity), nil
}
