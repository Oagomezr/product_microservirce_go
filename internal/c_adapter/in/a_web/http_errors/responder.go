package httpErrors

import (
	"encoding/json"
	"errors"
	"net/http"
	appErrors "product-microservice/internal/b_application/app_errors"
)

var statusByKind = map[appErrors.Kind]int{
	appErrors.KindValidation:     http.StatusBadRequest,
	appErrors.KindNotFound:       http.StatusNotFound,
	appErrors.KindAlreadyExists:  http.StatusConflict,
	appErrors.KindInternalServer: http.StatusInternalServerError,
}

func CustomError(w http.ResponseWriter, err error) {
	var appErr *appErrors.AppError
	if !errors.As(err, &appErr) {
		appErr = appErrors.InternalServer("internal server error")
	}

	status, ok := statusByKind[appErr.Kind()]
	if !ok {
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": appErr.Error()})
}

func ClientError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
