package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Create a new router
	mux := http.NewServeMux()

	handler := InitApplication()

	handler.routes(mux)

	// Start the server
	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
