package command

import (
	"encoding/json"
	"net/http"
	"product-microservice/internal/b_application/port/in"
	"product-microservice/internal/c_adapter/in/a_web/command/dto"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"

	"github.com/google/uuid"
)

// UpdateProductHandler manage the http request to update a product
type UpdateProductHandler struct {
	// CQRS, we just depend on the in input ports
	updaterUserCase in.ProductUpdater
}

// NewUpdateProductHandler is the manual constructor
func NewUpdateProductHandler(updaterUserCase in.ProductUpdater) *UpdateProductHandler {
	return &UpdateProductHandler{updaterUserCase: updaterUserCase}
}

// Handle updates a product
func (h *UpdateProductHandler) Handle(w http.ResponseWriter, r *http.Request) {
	//1. Validate the request method to be PATCH
	if r.Method != http.MethodPatch {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the request body (JSON) into a domain model
	var product dto.ProductPatchDTO
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		httpErrors.ClientError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	productID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpErrors.ClientError(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	productDomain := dto.ToDomainPatch(productID, product)

	// 3. Extract the context from the http request (r.context)
	// This is the life cycle of the request which will travel through our services postgres and kafka
	ctx := r.Context()

	// 4. Call the application service to update the product
	err = h.updaterUserCase.Execute(ctx, &productDomain)
	if err != nil {
		httpErrors.CustomError(w, err)
		return
	}

	// 5. Return the success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Product updated successfully"})
}
