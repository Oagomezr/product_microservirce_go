package command

import (
	"encoding/json"
	"log"
	"net/http"
	"product-microservice/internal/b_application/port/in"
	"product-microservice/internal/c_adapter/in/a_web/command/dto"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"

	"github.com/google/uuid"
)

// CreateProductHandler manage the http request to create a product
type CreateProductHandler struct {
	// CQRS, we just depend on the in input ports
	creatorUserCase in.ProductCreator
}

// NewCreateProductHandler is the manual constructor
func NewCreateProductHandler(creatorUserCase in.ProductCreator) *CreateProductHandler {
	return &CreateProductHandler{creatorUserCase: creatorUserCase}
}

// Handle creates a new product
func (h *CreateProductHandler) Handle(w http.ResponseWriter, r *http.Request) {
	//1. Validate the request method to be POST
	if r.Method != http.MethodPost {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the request body (JSON) into a domain model
	var product dto.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		httpErrors.ClientError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	productDomain := dto.ToDomain(uuid.Nil, product)

	// 3. Extract the context from the http request (r.context)
	// This is the life cycle of the request which will travel through our services postgres and kafka
	ctx := r.Context()

	// 4. trigger the use case
	productID, err := h.creatorUserCase.Execute(ctx, &productDomain)
	if err != nil {
		log.Printf("create product: failed: %v", err)
		httpErrors.CustomError(w, err)
		return
	}
	log.Printf("create product: ok id=%s name=%q", productID, productDomain.Name)

	// 5. Return the response (HTTP 201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uuid.UUID{"product_id": productID})
}
