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
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, `{"error": "failed to encode JSON"}`, http.StatusInternalServerError)
		}
	}
}
