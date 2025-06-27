package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type GenericQueryHandler[R any] func(ctx context.Context, input map[string]string) (R, error)

// HandleGenericGet handles GET requests by parsing query parameters into a map
func HandleGenericGet[R any](handler GenericQueryHandler[R]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check method get
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(Response[R]{
				Status:  StatusError,
				Message: "Method not allowed",
			})
			return
		}
		// Set content type header
		w.Header().Set("Content-Type", "application/json")

		queries := map[string]string{}
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queries[key] = values[0]
			}
		}

		resp, err := handler(r.Context(), queries)

		var response Response[R]
		if err != nil {
			log.Println("Handler execution failed", err)
			w.WriteHeader(http.StatusInternalServerError)
			response = Response[R]{
				Status:  StatusError,
				Message: "Handler execution failed",
				Error:   err.Error(),
			}
		} else {
			w.WriteHeader(http.StatusOK)
			response = Response[R]{
				Status:  StatusSuccess,
				Message: "Operation completed successfully",
				Data:    resp,
			}
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, `{"error": "failed to encode JSON"}`, http.StatusInternalServerError)
		}
	}
}
