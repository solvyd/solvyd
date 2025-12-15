package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/solvyd/solvyd/api-server/internal/database"
)

// HealthCheck handles health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// ReadinessCheck returns a handler that checks database connectivity
func ReadinessCheck(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "not ready",
				"error":  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
	}
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SendError sends a JSON error response
func SendError(w http.ResponseWriter, code int, err error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errMsg,
		Message: message,
		Code:    code,
	})
}

// SendJSON sends a JSON response
func SendJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
