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

// 1. Define the service struct
// This struct "keeps" the tools (outports) needed
type CreateProduct struct {
	saver  out.ProductSaver
	outbox out.OutboxSaver
	db     *sql.DB
}

// 2. constructor function (manual constructor on go)
// This make sure that no one can create the service without the tools/dependencies (outports)
func NewCreateProduct(saver out.ProductSaver, outbox out.OutboxSaver, db *sql.DB) *CreateProduct {
	return &CreateProduct{
		saver:  saver,
		outbox: outbox,
		db:     db,
	}
}

// 3. Implement the service logic(execute method)
// this method is the implementation of product creator (port/in)
func (s *CreateProduct) Execute(ctx context.Context, product *model.Product) (uuid.UUID, error) {
	// 1. Check if the product already exists
	if product.ID != uuid.Nil {
		return uuid.Nil, appErrors.Validation("product ID must be empty")
	}

	// 2. validate the product
	if err := product.Validate(); err != nil {
		return uuid.Nil, appErrors.Validation(err.Error())
	}

	// 3. begin the transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, appErrors.InternalServer(err.Error())
	}
	defer tx.Rollback()

	// 4. create the product id and timestamps
	product.ID = uuid.New()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// 5. marshal the product to json
	payload, err := json.Marshal(product)
	if err != nil {
		return uuid.Nil, appErrors.InternalServer(err.Error())
	}

	// 6. create the outbox event
	outboxEvent := &model.OutboxEvent{
		ID:          uuid.New(),
		AggregateID: product.ID,
		EventType:   "product_created",
		Payload:     payload,
		CreatedAt:   time.Now(),
	}

	// 7. save the outbox event
	if err := s.outbox.Save(ctx, tx, outboxEvent); err != nil {
		return uuid.Nil, appErrors.InternalServer(err.Error())
	}

	// 8. save the product
	err = s.saver.Save(ctx, tx, product)
	if err != nil {
		return uuid.Nil, appErrors.InternalServer(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, appErrors.InternalServer(err.Error())
	}

	return product.ID, nil
}
