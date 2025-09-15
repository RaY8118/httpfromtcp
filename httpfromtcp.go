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
type Handler interface {
	ServeHTTP(w *response.Writer, r *request.Request)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as HTTP handlers.
// If f is a function with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(w *response.Writer, r *request.Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w *response.Writer, r *request.Request) {
	f(w, r)
}

// ListenAndServe starts an HTTP server with a given address and handler.
func ListenAndServe(addr string, handler Handler) error {
	var port uint16
	if _, err := fmt.Sscanf(addr, ":%d", &port); err != nil {
		return fmt.Errorf("invalid address format: %s", addr)
	}

	// We pass the handler's ServeHTTP method to the internal server.
	s, err := server.Serve(port, handler.ServeHTTP)
	if err != nil {
		return err
	}
	defer s.Close()
	log.Printf("Server started on port %d", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Server gracefully stopped")

	return nil
}