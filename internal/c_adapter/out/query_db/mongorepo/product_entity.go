package mongorepo

import (
	"product-microservice/internal/a_domain/model"
	"time"

	"github.com/google/uuid"
)

type ProductEntity struct {
	ID               uuid.UUID `bson:"id"`
	Name             string    `bson:"name"`
	Price            float64   `bson:"price"`
	Stock            uint      `bson:"stock"`
	Description      string    `bson:"description"`
	ShortDescription string    `bson:"short_description"`
	Hidden           bool      `bson:"hidden"`
	CreatedAt        time.Time `bson:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at"`
	ImagesURLs       []string  `bson:"images_urls"`
	CategoryID       uuid.UUID `bson:"category_id"`
}

func (e *ProductEntity) toDomain() *model.Product {
	return &model.Product{
		ID:               e.ID,
		Name:             e.Name,
		Price:            e.Price,
		Stock:            e.Stock,
		Description:      e.Description,
		ShortDescription: e.ShortDescription,
		Hidden:           e.Hidden,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
		ImagesURLs:       e.ImagesURLs,
		CategoryID:       e.CategoryID,
	}
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
