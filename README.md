# HTTP Server from Scratch in Go

This project is a hands-on exploration into the fundamentals of the HTTP protocol, built from the ground up in Go using only raw TCP sockets. What started as a simple TCP listener evolved into a mini, `net/http`-like web library, complete with a request router and dynamic URL parameters.

The primary goal of this project was not to replace Go's excellent standard library, but to deconstruct and understand the magic that happens under the hood of a real web server.

## Features

*   **Built from Raw TCP Sockets:** The server is built on `net.Listener`, handling raw TCP connections to parse and respond to HTTP requests.
*   **Custom HTTP Parser:** Manually parses request lines, headers, and bodies according to the HTTP/1.1 protocol specification.
*   **Dynamic Request Router (Mux):** A `net/http`-style multiplexer that can route requests based on both the HTTP method (GET, POST, etc.) and the URL path.
*   **URL Parameter Support:** The router can extract dynamic parts from a URL, such as `/users/{id}`.
*   **Reusable Library Structure:** The project is structured as a Go library, with a clean public API and a separate `examples` directory to showcase its usage.

## Project Structure

```
.
├── httpfromtcp.go      # The public API of the library (like net/http)
├── internal/           # All the private, internal logic for the library
│   ├── headers/
│   ├── mux/
│   ├── request/
│   ├── response/
│   └── server/
├── examples/
│   └── simple-server/  # An example application using the library
│       ├── main.go
│       └── example_handlers.go
└── go.mod
```

## Getting Started

To run the example server, clone the repository and run the following command from the root directory:

```bash
go run ./examples/simple-server
```

The server will start on `localhost:42069`.

### Example Requests

You can use `curl` to interact with the running server:

**GET a simple response:**
```bash
curl http://localhost:42069/
```

**GET a route with a URL parameter:**
```bash
curl http://localhost:42069/hello/Parth
# Expected Output: Hello, Parth
```

**POST a new message:**
```bash
curl -X POST -d "This is a test message" http://localhost:42069/messages
# Expected Output: Message created successfully: This is a test message
```

## Library API (A Quick Look)

Using the library is designed to be simple and familiar to anyone who has used Go's `net/http` package. The example server's `main.go` is a great reference:

```go
package main

import (
	"log"

	httpfromtcp "ray8118/httpfromtcp"
	"ray8118/httpfromtcp/internal/mux"
)

func main() {
	// Create a new mux
	m := mux.NewMux()

	// Register handlers
	registerExampleHandlers(m)

	log.Println("Starting server on :42069")

	// Use the library's ListenAndServe function
	httpfromtcp.ListenAndServe(":42069", m)
}
```

## Roadmap

This project is a continuous learning exercise. The next planned features are outlined in `IMPROVEMENTS.md` and include:

*   A flexible Middleware pattern
*   Enhanced response helpers (e.g., for JSON)
*   Static file serving

## Inspiration

This project was inspired by a video from ThePrimeagen on YouTube, which is based on a course from [boot.dev](https://boot.dev). It was an incredible learning experience to follow along and then extend the project with new features.