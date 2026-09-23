package product

import (
	"product-microservice/internal/a_domain/model"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProductEntity struct {
	ID               uuid.UUID
	Name             string
	Price            float64
	Stock            uint
	Description      string
	ShortDescription string
	Hidden           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ImagesURLs       pq.StringArray
	CategoryID       uuid.UUID
}

func fromDomain(product *model.Product) *ProductEntity {
	return &ProductEntity{
		ID:               product.ID,
		Name:             product.Name,
		Price:            product.Price,
		Stock:            product.Stock,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Hidden:           product.Hidden,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
		ImagesURLs:       product.ImagesURLs,
		CategoryID:       product.CategoryID,
	}
}

func toDomain(productEntity *ProductEntity) *model.Product {
	return &model.Product{
		ID:               productEntity.ID,
		Name:             productEntity.Name,
		Price:            productEntity.Price,
		Stock:            productEntity.Stock,
		Description:      productEntity.Description,
		ShortDescription: productEntity.ShortDescription,
		Hidden:           productEntity.Hidden,
		CreatedAt:        productEntity.CreatedAt,
		UpdatedAt:        productEntity.UpdatedAt,
		ImagesURLs:       productEntity.ImagesURLs,
		CategoryID:       productEntity.CategoryID,
	}
}
