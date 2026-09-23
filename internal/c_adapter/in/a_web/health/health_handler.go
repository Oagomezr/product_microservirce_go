package health

import (
	"database/sql"

	"net/http"
	httpErrors "product-microservice/internal/c_adapter/in/a_web/http_errors"

	"context"
	"time"

	"encoding/json"

	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	pg    *sql.DB
	mongo *mongo.Client
}

func NewHealthHandler(pg *sql.DB, mongo *mongo.Client) *HealthHandler {
	return &HealthHandler{pg: pg, mongo: mongo}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {

	//1. Validate the request method to be GET
	if r.Method != http.MethodGet {
		httpErrors.ClientError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := ""
	code := http.StatusOK

	if err := h.pg.PingContext(ctx); err != nil {
		status += "Postgres not available |"
		code = http.StatusServiceUnavailable
	}

	if err := h.mongo.Ping(ctx, nil); err != nil {
		status += "Mongo not available |"
		code = http.StatusServiceUnavailable
	}

	if status == "" {
		status = "All services are healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}
