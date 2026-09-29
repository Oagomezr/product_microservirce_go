// Domain tests for Product — validation rules live here, not in the service layer.
package model

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestProduct_Validate(t *testing.T) {
	t.Run("rejects empty product", func(t *testing.T) {
		err := (&Product{}).Validate()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Error(), "name is required") {
			t.Fatalf("expected name error, got: %v", err)
		}
	})

	t.Run("accepts product with empty images slice", func(t *testing.T) {
		// nil slice fails; empty non-nil slice is allowed for now.
		p := &Product{
			Name:             "Pasta dental",
			Price:            12.5,
			Stock:            10,
			Description:      "Desc",
			ShortDescription: "Short",
			ImagesURLs:       []string{},
			CategoryID:       uuid.New(),
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("expected valid product, got: %v", err)
		}
	})

	t.Run("rejects missing images field", func(t *testing.T) {
		p := &Product{
			Name:             "Pasta dental",
			Price:            12.5,
			Stock:            10,
			Description:      "Desc",
			ShortDescription: "Short",
			CategoryID:       uuid.New(),
		}
		if err := p.Validate(); err == nil {
			t.Fatal("expected validation error for nil images_urls")
		}
	})
}
