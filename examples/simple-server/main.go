package main

import (
	"log"

	"ray8118/httpfromtcp"
	"ray8118/httpfromtcp/internal/mux"
)

const addr = ":42069"

func main() {
	// Create a new mux from our library
	m := mux.NewMux()

	// Register the example handlers (from example_handlers.go)
	registerExampleHandlers(m)

	log.Printf("Starting server on %s", addr)

	// Chain the middleware to the mux's ServeHTTP method.
	chainedHandler := mux.Chain(m.ServeHTTP, mux.LoggingMiddleware)

	// Convert the resulting HandlerFunc back into a Handler that ListenAndServe can accept.
	err := httpfromtcp.ListenAndServe(addr, httpfromtcp.HandlerFunc(chainedHandler))
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

