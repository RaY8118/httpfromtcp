package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
)

// Handler is the function signature for a request handler. It takes a response writer
// and a pointer to the parsed request.
type Handler func(w *response.Writer, req *request.Request)

// Server represents our HTTP server.
type Server struct {
	closed  bool
	handler Handler
}

// runConnection is responsible for handling a single TCP connection.
func runConnection(s *Server, conn io.ReadWriteCloser) {
	// Ensure the connection is closed when this function exits.
	defer conn.Close()

	// Create a response writer that writes back to the connection.
	responseWriter := response.NewWriter(conn)

	// Use the request parser to read from the connection and build a request object.
	r, err := request.RequestFromReader(conn)
	if err != nil {
		// If parsing fails, send a 400 Bad Request response.
		// A more robust server might log this error.
		log.Printf("Failed to parse request: %v", err)
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(*response.GetDefaultHeaders(0))
		return
	}

	// The request was parsed successfully. Call the main handler to generate a response.
	s.handler(responseWriter, r)
}

// runServer is the main loop that accepts incoming TCP connections.
func runServer(s *Server, listener net.Listener) {
	// Loop indefinitely, waiting for new connections.
	for {
		// Block until a new connection is received.
		conn, err := listener.Accept()
		if err != nil {
			// If the server has been closed, we can expect an error here, so we just exit.
			if s.closed {
				log.Println("Accept loop closed.")
				return
			}
			log.Printf("Failed to accept connection: %v", err)
			return
		}

		// Handle each new connection in its own goroutine.
		// This allows the server to handle multiple requests concurrently.
		go runConnection(s, conn)
	}
}

// Serve is the entry point for starting the server. It sets up the TCP listener
// and starts the main accept loop in a new goroutine.
func Serve(port uint16, handler Handler) (*Server, error) {
	// Start listening for TCP connections on the given port.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{closed: false, handler: handler}

	// Start the main server loop in a separate goroutine so that Serve can return immediately.
	go runServer(server, listener)

	return server, nil
}

// Close signals the server to stop accepting new connections.
func (s *Server) Close() error {
	s.closed = true
	return nil
}
