package httpfromtcp

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
	"ray8118/httpfromtcp/internal/server"
	"syscall"
)

// Handler is an interface that objects can implement to be a request handler.
// It is modeled directly after Go's standard `net/http.Handler` interface.
// Any object that implements this interface can be used as a request handler
// in the ListenAndServe function.
//
// The ServeHTTP method is called for each incoming request and is responsible
// for writing the response.
type Handler interface {
	ServeHTTP(w *response.Writer, r *request.Request)
}

// ListenAndServe starts an HTTP server with a given address and handler.
// It is a blocking call that will run until the program is interrupted (e.g., by Ctrl+C).
// This function is the primary entry point for the library, designed to feel
// familiar to users of Go's standard `net/http.ListenAndServe`.
func ListenAndServe(addr string, handler Handler) error {
	// For now, we parse the port from the address string in a simple way.
	// A more robust implementation would handle hostnames as well (e.g., "localhost:8080").
	var port uint16
	if _, err := fmt.Sscanf(addr, ":%d", &port); err != nil {
		return fmt.Errorf("invalid address format: %s", addr)
	}

	// server.Serve is the internal function that sets up the TCP listener.
	// We pass it the port and the handler's ServeHTTP method.
	s, err := server.Serve(port, handler.ServeHTTP)
	if err != nil {
		return err
	}
	defer s.Close()
	log.Printf("Server started on port %d", port)

	// Set up a channel to listen for OS signals (SIGINT, SIGTERM).
	// This allows for a graceful shutdown of the server.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	<-sigChan
	log.Println("Server gracefully stopped")

	return nil
}
