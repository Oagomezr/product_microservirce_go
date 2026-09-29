// HTTP adapter tests for POST /products.
//
// package command (same as the handler) — tests the web layer in isolation.
// httptest.NewRecorder simulates the response; no real server or port needed
// (similar to MockMvc in Spring Boot, but built into the standard library).
//
// stubProductCreator replaces the use case — we only assert HTTP status and JSON body.
package command

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"

	"github.com/google/uuid"
)

// stubProductCreator implements in.ProductCreator — returns a fixed id or error.
type stubProductCreator struct {
	id  uuid.UUID
	err error
}

func (s *stubProductCreator) Execute(_ context.Context, _ *model.Product) (uuid.UUID, error) {
	return s.id, s.err
}

// TestCreateProductHandler_Handle_ReturnsProductID verifies 201 + product_id in JSON.
func TestCreateProductHandler_Handle_ReturnsProductID(t *testing.T) {
	expectedID := uuid.New()
	handler := NewCreateProductHandler(&stubProductCreator{id: expectedID})

	body := bytes.NewBufferString(`{
		"name": "Pasta dental",
		"price": 12.5,
		"stock": 10,
		"description": "Pasta con fluor",
		"short_description": "Pasta",
		"category_id": "550e8400-e29b-41d4-a716-446655440000"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/products", body)
	respRec := httptest.NewRecorder()

	handler.Handle(respRec, req)

	if respRec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", respRec.Code, http.StatusCreated)
	}

	var response map[string]uuid.UUID
	if err := json.NewDecoder(respRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["product_id"] != expectedID {
		t.Fatalf("product_id = %s, want %s", response["product_id"], expectedID)
	}
}

// TestCreateProductHandler_Handle_MethodNotAllowed verifies GET on POST route → 405.
func TestCreateProductHandler_Handle_MethodNotAllowed(t *testing.T) {
	handler := NewCreateProductHandler(&stubProductCreator{})
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	handler.Handle(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

// TestCreateProductHandler_Handle_ValidationError verifies use case errors map to 400.
func TestCreateProductHandler_Handle_ValidationError(t *testing.T) {
	handler := NewCreateProductHandler(&stubProductCreator{
		err: appErrors.Validation("name is required"),
	})

	body := bytes.NewBufferString(`{"name":"","price":1,"stock":1,"description":"d","short_description":"s","category_id":"550e8400-e29b-41d4-a716-446655440000"}`)
	req := httptest.NewRequest(http.MethodPost, "/products", body)
	rec := httptest.NewRecorder()

	handler.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
