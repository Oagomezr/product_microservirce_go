package query

import (
	"encoding/json"
	"net/http"
	"product-microservice/internal/b_application/port/in"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"

	"github.com/google/uuid"
)

// FindProductHandler manage the http request to find a product
type FindProductHandler struct {
	finderUserCase in.ProductFinder
}

// NewFindProductHandler is the manual constructor
func NewFindProductHandler(finderUserCase in.ProductFinder) *FindProductHandler {
	return &FindProductHandler{finderUserCase: finderUserCase}
}

// Handle finds a product
func (h *FindProductHandler) Handle(w http.ResponseWriter, r *http.Request) {
	//1. Validate the request method to be GET
	if r.Method != http.MethodGet {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract the product ID from the request URL
	productIDStr := r.PathValue("id")
	if productIDStr == "" {
		httpErrors.ClientError(w, "Product ID is required", http.StatusBadRequest)
		return
	}
	productID, err := uuid.Parse(productIDStr)

	if err != nil {
		httpErrors.ClientError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// 3. Extract the context from the http request (r.context)
	// This is the life cycle of the request which will travel through our services postgres and kafka
	ctx := r.Context()

	// 4. Call the application service to find the product
	product, err := h.finderUserCase.Execute(ctx, productID)
	if err != nil {
		httpErrors.CustomError(w, err)
		return
	}

	// 5. Return the success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}
