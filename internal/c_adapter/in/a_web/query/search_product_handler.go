package query

import (
	"encoding/json"
	"net/http"
	"product-microservice/internal/b_application/port/in"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"
)

// SearchProductHandler manage the http request to search a product
type SearchProductHandler struct {
	searcherUserCase in.ProductSearcher
}

// NewSearchProductHandler is the manual constructor
func NewSearchProductHandler(searcherUserCase in.ProductSearcher) *SearchProductHandler {
	return &SearchProductHandler{searcherUserCase: searcherUserCase}
}

// Handle searches a product
func (h *SearchProductHandler) Handle(w http.ResponseWriter, r *http.Request) {
	//1. Validate the request method to be GET
	if r.Method != http.MethodGet {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract the query parameters from the request URL
	query := r.URL.Query().Get("query")
	if len(query) < 3 {
		httpErrors.ClientError(w, "Query must be at least 3 characters long", http.StatusBadRequest)
		return
	}

	// 3. Extract the context from the http request (r.context)
	// This is the life cycle of the request which will travel through our services postgres and kafka
	ctx := r.Context()

	// 4. Call the application service to search the product
	products, err := h.searcherUserCase.Execute(ctx, query)
	if err != nil {
		httpErrors.CustomError(w, err)
		return
	}

	// 5. Return the success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}
