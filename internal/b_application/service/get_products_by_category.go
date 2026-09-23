package service

import (
	"context"
	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"
	"product-microservice/internal/b_application/port/out"

	"github.com/google/uuid"
)

type GetProductsByCategory struct {
	loader out.CategoryProductsLoader
}

func NewGetProductsByCategory(loader out.CategoryProductsLoader) *GetProductsByCategory {
	return &GetProductsByCategory{
		loader: loader,
	}
}

func (s *GetProductsByCategory) Execute(ctx context.Context, categoryID uuid.UUID) ([]*model.Product, error) {
	products, err := s.loader.LoadByCategory(ctx, categoryID)
	if err != nil {
		return nil, appErrors.InternalServer(err.Error())
	}
	return products, nil
}
