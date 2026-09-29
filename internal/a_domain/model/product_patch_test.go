// Domain tests for ProductPatch — IsEmpty() and Validate() belong to the domain layer.
package model

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestProductPatch_IsEmpty(t *testing.T) {
	t.Run("empty patch", func(t *testing.T) {
		patch := &ProductPatch{ID: uuid.New()}
		if !patch.IsEmpty() {
			t.Fatal("expected patch to be empty")
		}
	})

	t.Run("patch with name", func(t *testing.T) {
		name := "Pasta dental"
		patch := &ProductPatch{Name: &name}
		if patch.IsEmpty() {
			t.Fatal("expected patch not to be empty")
		}
	})
}

func TestProductPatch_Validate(t *testing.T) {
	t.Run("rejects blank name when field is sent", func(t *testing.T) {
		name := "   "
		err := (&ProductPatch{Name: &name}).Validate()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "name is required") {
			t.Fatalf("expected name error, got: %v", err)
		}
	})

	t.Run("rejects non-positive price when field is sent", func(t *testing.T) {
		price := 0.0
		err := (&ProductPatch{Price: &price}).Validate()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "price must be greater than 0") {
			t.Fatalf("expected price error, got: %v", err)
		}
	})

	t.Run("allows omitted fields", func(t *testing.T) {
		name := "Valid"
		err := (&ProductPatch{Name: &name}).Validate()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("rejects nil uuid category when field is sent", func(t *testing.T) {
		categoryID := uuid.Nil
		err := (&ProductPatch{CategoryID: &categoryID}).Validate()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "category ID is required") {
			t.Fatalf("expected category error, got: %v", err)
		}
	})
}
