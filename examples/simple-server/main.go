package main

import (
	"log"

	httpfromtcp "ray8118/httpfromtcp"
	"ray8118/httpfromtcp/internal/mux"
)

const addr = ":42069"

func main() {
	// Create a new mux from our library
	m := mux.NewMux()

	// Register the example handlers (from example_handlers.go)
	registerExampleHandlers(m)

	log.Printf("Starting server on %s", addr)

	// Use the new high-level ListenAndServe function from our library.
	// This is a blocking call.
	err := httpfromtcp.ListenAndServe(addr, m)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
