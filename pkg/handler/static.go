package handler

import (
	"log"
	"net/http"
)

type RenderStatic func(w http.ResponseWriter, r *http.Request) error

func HandleStatic(renderer RenderStatic) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderer(w, r); err != nil {
			log.Println("error rendering static file:", err)
		}
	}
}
