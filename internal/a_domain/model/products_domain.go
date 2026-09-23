package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID               uuid.UUID
	Name             string
	Price            float64
	Stock            uint
	Description      string
	ShortDescription string
	Hidden           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ImagesURLs       []string
	CategoryID       uuid.UUID
}

func (p *Product) Validate() error {
	concatErrors := strings.Builder{}
	if strings.TrimSpace(p.Name) == "" {
		concatErrors.WriteString("name is required\n")
	}
	if p.Price <= 0 {
		concatErrors.WriteString("price must be greater than 0\n")
	}
	if strings.TrimSpace(p.Description) == "" {
		concatErrors.WriteString("description is required\n")
	}
	if strings.TrimSpace(p.ShortDescription) == "" {
		concatErrors.WriteString("short description is required\n")
	}
	if p.ImagesURLs == nil {
		concatErrors.WriteString("images URLs are required\n")
	}
	if p.CategoryID == uuid.Nil {
		concatErrors.WriteString("category ID is required\n")
	}
	if concatErrors.Len() > 0 {
		return errors.New(concatErrors.String())
	}
	return nil
}
