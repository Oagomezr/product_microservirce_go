package model

import (
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
