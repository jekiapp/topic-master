package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type GenericQueryHandler[R any] func(ctx context.Context, input map[string]string) (R, error)

// QueryHandler handles GET requests by parsing query parameters into a map
func QueryHandler[R any](handler GenericQueryHandler[R]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		queries := map[string]string{}
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queries[key] = values[0]
			}
		}

		resp, err := handler(r.Context(), queries)
		w.Header().Set("Content-Type", "application/json")

		var response Response[R]
		if err != nil {
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
