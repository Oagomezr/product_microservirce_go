package model

import (
	"github.com/google/uuid"
)

type ProductPatch struct {
	ID               uuid.UUID
	Name             *string
	Price            *float64
	Stock            *uint
	Description      *string
	ShortDescription *string
	Hidden           *bool
	ImagesURLs       *[]string
	CategoryID       *uuid.UUID
}

func (p *ProductPatch) IsEmpty() bool {
	return p.Name == nil &&
		p.Price == nil &&
		p.Stock == nil &&
		p.Description == nil &&
		p.ShortDescription == nil &&
		p.Hidden == nil &&
		p.ImagesURLs == nil &&
		p.CategoryID == nil
}
