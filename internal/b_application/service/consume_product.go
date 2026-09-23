package service

import (
	"context"
	"product-microservice/internal/a_domain/model"
	"product-microservice/internal/b_application/port/out"

	"github.com/google/uuid"
)

type ConsumeProduct struct {
	saver   out.ProductProjectionSaver
	deleter out.ProductProjectionDeleter
}

func NewConsumeProduct(saver out.ProductProjectionSaver, deleter out.ProductProjectionDeleter) *ConsumeProduct {
	return &ConsumeProduct{
		saver:   saver,
		deleter: deleter,
	}
}

func (s *ConsumeProduct) ProjectUpsert(ctx context.Context, product *model.Product) error {
	return s.saver.Upsert(ctx, product)
}

func (s *ConsumeProduct) ProjectDelete(ctx context.Context, id uuid.UUID) error {
	return s.deleter.Delete(ctx, id)
}
