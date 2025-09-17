package mux

import (
	"log"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
	"strings"
	"time"
)

// HandlerFunc defines the signature for handlers that our Mux will use.
type HandlerFunc func(w *response.Writer, r *request.Request)

// Middleware is a function that takes a handler and returns a new handler.
// This allows for chaining, where each middleware can perform some action before or after calling the next handler in the chain.
type Middleware func(HandlerFunc) HandlerFunc

// route holds the information for a single registered route.
// This includes the HTTP method, the handler function, and the path
// broken down into its component parts.
type route struct {
	method  string
	handler HandlerFunc
	parts   []string // e.g., "/users/{id}" becomes ["users", "{id}"]
}

// Mux is a request router (or multiplexer). It matches incoming requests
// against a list of registered patterns and calls the handler for the
// pattern that matches the URL.
type Mux struct {
	// We use a slice of routes instead of a map to allow for dynamic matching.
	// A map would only allow for exact path matches.
	routes []*route
}

// NewMux creates and returns a new Mux.
func NewMux() *Mux {
	return &Mux{
		routes: make([]*route, 0),
	}
}

func Chain(h HandlerFunc, mws ...Middleware) HandlerFunc {
	// Start with the final handler
	handler := h

	// Loop backwards through the middlewares, wrapping the current handler with each one
	// This ensures they execute in hte correct order
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler

}

// LoggingMiddleware logs the details of each request
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w *response.Writer, r *request.Request) {
		start := time.Now()
		next(w, r)

		log.Printf("method=%s path=%s duration=%s", r.RequestLine.Method, r.RequestLine.RequestTarget, time.Since(start))
	}
}

// HandleFunc registers a new handler function for the given method and path.
func (m *Mux) HandleFunc(method, path string, handler HandlerFunc) {
	newRoute := &route{
		method:  method,
		handler: handler,
		// We trim the slashes and split the path so we can compare it part-by-part later.
		parts: strings.Split(strings.Trim(path, "/"), "/"),
	}
	m.routes = append(m.routes, newRoute)
}

// ServeHTTP is the main entry point for routing. It finds the correct handler
// for the request and calls it. If no handler is found, it returns a 404 Not Found error.
func (m *Mux) ServeHTTP(w *response.Writer, r *request.Request) {
	// Split the incoming request path into parts so we can compare it with our registered routes.
	requestParts := strings.Split(strings.Trim(r.RequestLine.RequestTarget, "/"), "/")

	// Loop through all registered routes to find a match.
	for _, route := range m.routes {
		// First, check if the HTTP method matches.
		if route.method != r.RequestLine.Method {
			continue
		}

		// Check if the number of path parts match. If not, this route can't possibly match.
		if len(route.parts) != len(requestParts) {
			continue
		}

		// Now, check each part of the path for a match.
		match := true
		params := make(map[string]string)
		for i, part := range route.parts {
			// Check if this part of the registered route is a dynamic parameter (e.g., "{id}").
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				// It is a parameter. Extract the name and store the corresponding value from the request path.
				paramName := strings.Trim(part, "{}")
				params[paramName] = requestParts[i]
			} else {
				// This is a static path part. It must match the request path part exactly.
				if part != requestParts[i] {
					match = false
					break
				}
			}
		}

		// If all parts matched, we've found our handler.
		if match {
			// Add the extracted parameters to the request object so the handler can access them.
			r.PathParams = params
			// Call the handler and stop searching.
			route.handler(w, r)
			return
		}
	}

	// If we looped through all routes and found no match, send a 404 Not Found response.
	w.WriteStatusLine(response.StatusNotFound)
	h := response.GetDefaultHeaders(0)
	body := []byte("404 Not Found")
	h.Replace("Content-length", "13")
	h.Replace("Content-type", "text/plain")
	w.WriteHeaders(*h)
	w.WriteBody(body)
}
