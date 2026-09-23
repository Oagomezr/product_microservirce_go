package service

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"
	"product-microservice/internal/b_application/port/out"

	"time"

	appErrors "product-microservice/internal/b_application/app_errors"

	"github.com/google/uuid"
)

type DeleteProduct struct {
	deleter out.ProductDeleter
	checker out.ProductChecker
	outbox  out.OutboxSaver
	db      *sql.DB
}

func NewDeleteProduct(deleter out.ProductDeleter, checker out.ProductChecker, outbox out.OutboxSaver, db *sql.DB) *DeleteProduct {
	return &DeleteProduct{
		deleter: deleter,
		checker: checker,
		outbox:  outbox,
		db:      db,
	}
}

func (s *DeleteProduct) Execute(ctx context.Context, id uuid.UUID) error {
	// 1. check if the product exists
	exist, err := s.checker.Exist(ctx, id)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}
	if !exist {
		return appErrors.NotFound("product not found")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}
	defer tx.Rollback()

	outboxEvent := &model.OutboxEvent{
		ID:          uuid.New(),
		AggregateID: id,
		EventType:   "product_deleted",
		Payload:     []byte(""),
		CreatedAt:   time.Now(),
	}

	if err := s.outbox.Save(ctx, tx, outboxEvent); err != nil {
		return appErrors.InternalServer(err.Error())
	}
	err = s.deleter.Delete(ctx, tx, id)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}
	if err := tx.Commit(); err != nil {
		return appErrors.InternalServer(err.Error())
	}
	return nil

}
