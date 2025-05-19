package handler

import "net/http"

type RenderStatic func(w http.ResponseWriter, r *http.Request) error

func HandleStatic(renderer RenderStatic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderer(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
