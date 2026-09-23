package dto

import (
	"product-microservice/internal/a_domain/model"

	"github.com/google/uuid"
)

func ToDomain(id uuid.UUID, product Product) model.Product {
	return model.Product{
		ID:               id,
		Name:             product.Name,
		Price:            product.Price,
		Stock:            product.Stock,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Hidden:           product.Hidden,
		ImagesURLs:       product.ImagesURLs,
		CategoryID:       product.CategoryID,
	}
}

func ToDomainPatch(id uuid.UUID, product ProductPatchDTO) model.ProductPatch {
	return model.ProductPatch{
		ID:               id,
		Name:             product.Name,
		Price:            product.Price,
		Stock:            product.Stock,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Hidden:           product.Hidden,
		ImagesURLs:       product.ImagesURLs,
		CategoryID:       product.CategoryID,
	}
}
