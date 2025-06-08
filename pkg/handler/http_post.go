package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ResponseStatus string

const (
	StatusSuccess ResponseStatus = "success"
	StatusError   ResponseStatus = "error"
)

type Response[O any] struct {
	Status  ResponseStatus `json:"status"`
	Message string         `json:"message"`
	Data    O              `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type GenericPostHandler[I any, O any] func(ctx context.Context, input I) (output O, err error)

func HandleGenericPost[I any, O any](handler GenericPostHandler[I, O]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// check method post
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(Response[O]{
				Status:  StatusError,
				Message: "Method not allowed",
			})
			return
		}
		// Set content type header
		w.Header().Set("Content-Type", "application/json")

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response[O]{
				Status:  StatusError,
				Message: "Failed to read request body",
				Error:   err.Error(),
			})
			return
		}
		defer r.Body.Close()

		// Parse request body
		data := new(I)
		if err := json.Unmarshal(body, data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response[O]{
				Status:  StatusError,
				Message: "Invalid JSON format",
				Error:   err.Error(),
			})
			return
		}

		// Validate input data
		if err := validate.Struct(data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response[O]{
				Status:  StatusError,
				Message: "Validation failed",
				Error:   err.Error(),
			})
			return
		}

		// Execute handler
		result, err := handler(r.Context(), *data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response[O]{
				Status:  StatusError,
				Message: "Handler execution failed",
				Data:    result,
				Error:   err.Error(),
			})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response[O]{
			Status:  StatusSuccess,
			Message: "Operation completed successfully",
			Data:    result,
		})
	}
}

// HandleGetPost returns a handler that dispatches to getHandler for GET and postHandler for POST, 405 otherwise.
func HandleGetPost[G any, I any, O any](getHandler GenericQueryHandler[G], postHandler GenericPostHandler[I, O]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HandleGenericGet(getHandler)(w, r)
		case http.MethodPost:
			HandleGenericPost(postHandler)(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
