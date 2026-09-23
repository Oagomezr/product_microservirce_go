package service

import (
	"context"
	"database/sql"
	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"
	"product-microservice/internal/b_application/port/out"
	"time"

	"encoding/json"

	"github.com/google/uuid"
)

type UpdateProduct struct {
	loader  out.ProductLoader
	updater out.ProductUpdater
	checker out.ProductChecker
	outbox  out.OutboxSaver
	db      *sql.DB
}

func NewUpdateProduct(loader out.ProductLoader, updater out.ProductUpdater, checker out.ProductChecker, outbox out.OutboxSaver, db *sql.DB) *UpdateProduct {
	return &UpdateProduct{
		loader:  loader,
		updater: updater,
		checker: checker,
		outbox:  outbox,
		db:      db,
	}
}

func (s *UpdateProduct) Execute(ctx context.Context, productPatch *model.ProductPatch) error {

	// 1. validate the product patch
	if productPatch.IsEmpty() {
		return appErrors.Validation("at least one field is required")
	}

	// 2. check if the product exists
	exists, err := s.checker.Exist(ctx, productPatch.ID)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}
	if !exists {
		return appErrors.NotFound("product not found")
	}

	// 3. load the product
	productLoaded, err := s.loader.Load(ctx, productPatch.ID)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}

	// 4. merge the product
	productMerged := s.mergeProduct(productLoaded, productPatch)

	// 5. begin the transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}
	defer tx.Rollback()

	// 6. create the outbox event
	payload, _ := json.Marshal(productMerged)

	outboxEvent := &model.OutboxEvent{
		ID:          uuid.New(),
		AggregateID: productPatch.ID,
		EventType:   "product_updated",
		Payload:     payload,
		CreatedAt:   time.Now(),
	}

	if err := s.outbox.Save(ctx, tx, outboxEvent); err != nil {
		return appErrors.InternalServer(err.Error())
	}

	// 7. update the product
	err = s.updater.Update(ctx, tx, productMerged)
	if err != nil {
		return appErrors.InternalServer(err.Error())
	}

	// 8. commit the transaction
	return tx.Commit()
}

func (s *UpdateProduct) mergeProduct(loaded *model.Product, patch *model.ProductPatch) *model.Product {

	merged := *loaded

	if patch.Name != nil {
		merged.Name = *patch.Name
	}
	if patch.Price != nil {
		merged.Price = *patch.Price
	}
	if patch.Stock != nil {
		merged.Stock = *patch.Stock
	}
	if patch.Description != nil {
		merged.Description = *patch.Description
	}
	if patch.ShortDescription != nil {
		merged.ShortDescription = *patch.ShortDescription
	}
	if patch.Hidden != nil {
		merged.Hidden = *patch.Hidden
	}
	if patch.ImagesURLs != nil {
		merged.ImagesURLs = *patch.ImagesURLs
	}
	if patch.CategoryID != nil {
		merged.CategoryID = *patch.CategoryID
	}

	merged.UpdatedAt = time.Now()

	return &merged
}
