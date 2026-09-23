package service

import (
	"context"
	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"
	"product-microservice/internal/b_application/port/out"
)

type SearchProduct struct {
	searcher out.ProductSearcher
}

func NewSearchProduct(searcher out.ProductSearcher) *SearchProduct {
	return &SearchProduct{
		searcher: searcher,
	}
}

func (s *SearchProduct) Execute(ctx context.Context, query string) ([]*model.Product, error) {
	products, err := s.searcher.Search(ctx, query)
	if err != nil {
		return nil, appErrors.InternalServer(err.Error())
	}
	return products, nil
}
