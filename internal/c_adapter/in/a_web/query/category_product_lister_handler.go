package query

import (
	"encoding/json"
	"net/http"
	"product-microservice/internal/b_application/port/in"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"

	"github.com/google/uuid"
)

// CategoryProductListerHandler manage the http request to list all products by category
type CategoryProductListerHandler struct {
	listerUserCase in.CategoryProductLister
}

// NewCategoryProductListerHandler is the manual constructor
func NewCategoryProductListerHandler(listerUserCase in.CategoryProductLister) *CategoryProductListerHandler {
	return &CategoryProductListerHandler{listerUserCase: listerUserCase}
}

// Handle lists all products by category
func (h *CategoryProductListerHandler) Handle(w http.ResponseWriter, r *http.Request) {
	//1. Validate the request method to be GET
	if r.Method != http.MethodGet {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract the category ID from the request URL
	categoryIDStr := r.URL.Query().Get("category_id")
	if categoryIDStr == "" {
		httpErrors.ClientError(w, "Category ID is required", http.StatusBadRequest)
		return
	}
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		httpErrors.ClientError(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	// 3. Extract the context from the http request (r.context)
	// This is the life cycle of the request which will travel through our services postgres and kafka
	ctx := r.Context()

	// 4. Call the application service to list all products by category
	products, err := h.listerUserCase.Execute(ctx, categoryID)
	if err != nil {
		httpErrors.CustomError(w, err)
		return
	}

	// 5. Return the success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}
