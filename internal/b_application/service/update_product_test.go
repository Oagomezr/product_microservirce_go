// Unit tests for the UpdateProduct use case (PATCH flow).
//
// mergeProduct is unexported but tested here because it holds core PATCH logic:
// only non-nil pointer fields from ProductPatch overwrite the loaded product.
// Execute tests cover validation and "not found" without needing a real database.
package service

import (
	"context"
	"testing"
	"time"

	"product-microservice/internal/a_domain/model"
	appErrors "product-microservice/internal/b_application/app_errors"

	"github.com/google/uuid"
)

// TestUpdateProduct_mergeProduct verifies partial updates: patched fields change,
// untouched fields (stock, created_at) stay the same, updated_at is refreshed.
func TestUpdateProduct_mergeProduct(t *testing.T) {
	svc := NewUpdateProduct(nil, nil, nil, nil, nil)

	loaded := &model.Product{
		ID:               uuid.New(),
		Name:             "Original",
		Price:            10,
		Stock:            5,
		Description:      "Desc",
		ShortDescription: "Short",
		Hidden:           false,
		ImagesURLs:       []string{"a.png"},
		CategoryID:       uuid.New(),
		CreatedAt:        time.Now().Add(-time.Hour),
		UpdatedAt:        time.Now().Add(-time.Hour),
	}

	newPrice := 19.99
	newName := "Updated"
	patch := &model.ProductPatch{
		ID:    loaded.ID,
		Name:  &newName,
		Price: &newPrice,
	}

	merged := svc.mergeProduct(loaded, patch)

	if merged.Name != newName {
		t.Fatalf("name = %q, want %q", merged.Name, newName)
	}
	if merged.Price != newPrice {
		t.Fatalf("price = %v, want %v", merged.Price, newPrice)
	}
	if merged.Stock != loaded.Stock {
		t.Fatalf("stock should remain unchanged")
	}
	if merged.CreatedAt != loaded.CreatedAt {
		t.Fatalf("created_at should remain unchanged")
	}
	if !merged.UpdatedAt.After(loaded.UpdatedAt) {
		t.Fatalf("updated_at should be refreshed")
	}
}

// TestUpdateProduct_Execute_EmptyPatch ensures PATCH with no fields returns 400.
func TestUpdateProduct_Execute_EmptyPatch(t *testing.T) {
	svc := NewUpdateProduct(nil, nil, nil, nil, nil)

	err := svc.Execute(context.Background(), &model.ProductPatch{ID: uuid.New()})
	assertAppErrorKind(t, err, appErrors.KindValidation)
}

// TestUpdateProduct_Execute_InvalidPatch ensures domain patch validation maps to 400.
func TestUpdateProduct_Execute_InvalidPatch(t *testing.T) {
	svc := NewUpdateProduct(nil, nil, nil, nil, nil)

	name := ""
	err := svc.Execute(context.Background(), &model.ProductPatch{
		ID:   uuid.New(),
		Name: &name,
	})
	assertAppErrorKind(t, err, appErrors.KindValidation)
}

// TestUpdateProduct_Execute_NotFound ensures updating a missing product returns 404.
func TestUpdateProduct_Execute_NotFound(t *testing.T) {
	checker := &stubProductChecker{exists: false}
	svc := NewUpdateProduct(nil, nil, checker, nil, nil)

	name := "Updated"
	err := svc.Execute(context.Background(), &model.ProductPatch{
		ID:   uuid.New(),
		Name: &name,
	})
	assertAppErrorKind(t, err, appErrors.KindNotFound)
}

// stubProductChecker implements out.ProductChecker for tests only.
type stubProductChecker struct {
	exists bool
	err    error
}

func (s *stubProductChecker) Exist(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.exists, s.err
}
