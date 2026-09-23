package service

import (
	"context"
	"product-microservice/internal/a_domain/model"
	"product-microservice/internal/b_application/port/out"

	appErrors "product-microservice/internal/b_application/app_errors"

	"github.com/google/uuid"
)

type GetProductById struct {
	checker out.ProductChecker
	loader  out.ProductLoader
}

func NewGetProductById(checker out.ProductChecker, loader out.ProductLoader) *GetProductById {
	return &GetProductById{
		checker: checker,
		loader:  loader,
	}
}

func (s *GetProductById) Execute(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	// 1. check if the product exists
	exist, err := s.checker.Exist(ctx, id)
	if err != nil {
		return nil, appErrors.InternalServer(err.Error())
	}
	if !exist {
		return nil, appErrors.NotFound("product not found")
	}
	// 2. load the product
	product, err := s.loader.Load(ctx, id)
	if err != nil {
		return nil, appErrors.InternalServer(err.Error())
	}
	return product, nil
}
