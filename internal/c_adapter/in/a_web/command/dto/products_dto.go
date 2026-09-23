package dto

import (
	"github.com/google/uuid"
)

type Product struct {
	Name             string    `json:"name"`
	Price            float64   `json:"price"`
	Stock            uint      `json:"stock"`
	Description      string    `json:"description"`
	ShortDescription string    `json:"short_description"`
	Hidden           bool      `json:"hidden"`
	ImagesURLs       []string  `json:"images_urls"`
	CategoryID       uuid.UUID `json:"category_id"`
}
