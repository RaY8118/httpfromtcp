# HTTP Server from Scratch in Go

[![Go Reference](https://pkg.go.dev/badge/ray8118/httpfromtcp.svg)](https://pkg.go.dev/ray8118/httpfromtcp)

This project is a hands-on exploration into the fundamentals of the HTTP protocol, built from the ground up in Go using only raw TCP sockets. What started as a simple TCP listener evolved into a mini, `net/http`-like web library.

The primary goal of this project was not to replace Go's excellent standard library, but to deconstruct and understand the magic that happens under the hood of a real web server.

## Features

*   **HTTP Server from Scratch:** Built on `net.Listener`, handling raw TCP connections to parse and respond to HTTP/1.1 requests.
*   **Custom Request Parser:** Manually parses request lines, headers, and bodies.
*   **Request Router (Mux):** A `net/http`-style multiplexer that routes requests based on method and URL path.
*   **Advanced Routing:** Supports dynamic URL parameters (e.g., `/users/{id}`) and query string parsing.
*   **Middleware:** A flexible middleware pattern for chaining functions to process requests, perfect for logging, auth, etc.
*   **Response Helpers:** Includes helpers like `JSON()` for easy JSON responses and `Respond200`, `Respond400`, etc., for standard HTTP responses.
*   **Static File Serving:** Serve static files and directories.
*   **Reusable Library Structure:** The project is structured as a Go library with a clean public API.

## Installation

To use this library in your own project, you can install it with `go get`:
```bash
go get ray8118/httpfromtcp@v0.2.0
```

## Usage

Here is a complete example of how to create a simple API server using the library.

```go
package main

import (
	"fmt"
	"log"

	"ray8118/httpfromtcp"
	"ray8118/httpfromtcp/internal/mux"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
	"ray8118/httpfromtcp/internal/static"
)

// Define a handler for your route
func helloHandler(w *response.Writer, r *request.Request) {
	// Use the JSON helper to send a response
	name, ok := r.PathParams["name"]
	if !ok {
		name = "World"
	}
	
	data := map[string]string{"message": fmt.Sprintf("Hello, %s!", name)}
	w.JSON(200, data)
}

func main() {
	// 1. Create a new router
	m := mux.NewMux()

	// 2. Register your handlers
	m.HandleFunc("GET", "/hello", helloHandler)
	m.HandleFunc("GET", "/hello/{name}", helloHandler)
	m.HandleFunc("GET", "/static/", static.Static)


	// 3. Create a middleware chain
	// The LoggingMiddleware will log every request
	handlerChain := mux.Chain(m.ServeHTTP, mux.LoggingMiddleware)

	fmt.Println("Starting server on :8080")

	// 4. Start the server
	// We wrap our chained handler in httpfromtcp.HandlerFunc to satisfy the interface
	err := httpfromtcp.ListenAndServe(":8080", httpfromtcp.HandlerFunc(handlerChain))
	if err != nil {
		log.Fatal(err)
	}
}
```

## Project Structure

```
.
├── httpfromtcp.go      # The public API of the library (like net/http)
├── internal/           # All the private, internal logic for the library
│   ├── headers/
│   ├── mux/
│   ├── request/
│   ├── response/
│   ├── server/
│   └── static/
├── examples/
│   └── simple-server/  # An example application using the library
└── go.mod
```

## Roadmap

This project is a continuous learning exercise. The next planned features are outlined in `IMPROVEMENTS.md` and include:

*   Configuration and Graceful Shutdown

## Inspiration

This project was inspired by a video from ThePrimeagen on YouTube, which is based on a course from [boot.dev](https://boot.dev). It was an incredible learning experience to follow along and then extend the project with new features.