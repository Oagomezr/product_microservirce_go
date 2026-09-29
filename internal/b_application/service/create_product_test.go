// Unit tests for the CreateProduct use case (orchestration + transaction wiring).
// Domain validation is tested in a_domain/model/products_domain_test.go.
package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func validProduct() *model.Product {
	return &model.Product{
		Name:             "Pasta dental",
		Price:            12.5,
		Stock:            10,
		Description:      "Pasta con fluor",
		ShortDescription: "Pasta",
		ImagesURLs:       []string{},
		CategoryID:       uuid.New(),
	}
}

func TestCreateProduct_Execute_RejectsPresetID(t *testing.T) {
	svc := NewCreateProduct(nil, nil, nil)
	product := validProduct()
	product.ID = uuid.New()

	_, err := svc.Execute(context.Background(), product)
	assertAppErrorKind(t, err, appErrors.KindValidation)
}

func TestCreateProduct_Execute_RejectsInvalidProduct(t *testing.T) {
	svc := NewCreateProduct(nil, nil, nil)

	_, err := svc.Execute(context.Background(), &model.Product{})
	assertAppErrorKind(t, err, appErrors.KindValidation)
}

func TestCreateProduct_Execute_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	saver := &recordingProductSaver{}
	outbox := &recordingOutboxSaver{}

	svCreateProduct := NewCreateProduct(saver, outbox, db)
	product := validProduct()

	id, err := svCreateProduct.Execute(context.Background(), product)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("expected generated product id")
	}
	if saver.product == nil {
		t.Fatal("expected product to be saved")
	}
	if saver.product.ID != id {
		t.Fatalf("saved product id mismatch: got %s want %s", saver.product.ID, id)
	}
	if outbox.event == nil {
		t.Fatal("expected outbox event to be saved")
	}
	if outbox.event.AggregateID != id {
		t.Fatalf("outbox aggregate id mismatch: got %s want %s", outbox.event.AggregateID, id)
	}
	if outbox.event.EventType != "product_created" {
		t.Fatalf("unexpected event type: %s", outbox.event.EventType)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type recordingProductSaver struct {
	product *model.Product
}

func (r *recordingProductSaver) Save(_ context.Context, _ *sql.Tx, product *model.Product) error {
	r.product = product
	return nil
}

type recordingOutboxSaver struct {
	event *model.OutboxEvent
}

func (r *recordingOutboxSaver) Save(_ context.Context, _ *sql.Tx, event *model.OutboxEvent) error {
	r.event = event
	return nil
}

func assertAppErrorKind(t *testing.T, err error, kind appErrors.Kind) {
	t.Helper()

	var appErr *appErrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	if appErr.Kind() != kind {
		t.Fatalf("expected kind %v, got %v", kind, appErr.Kind())
	}
}
