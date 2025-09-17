# HTTP Server Enhancement Guide

This guide outlines the steps to evolve your custom HTTP server into a more feature-rich and structured framework, similar to Go's `net/http`.

## Step 1: Implementing a Router (Mux)

Your current server has a simple `if/else` block for routing. A router, or multiplexer (mux), will make this much cleaner and more scalable.

**Goal:** Create a router that can map different URL paths and HTTP methods to specific handler functions.

**How-to:**

1.  **Create a `Mux` struct:**
    *   This struct will hold the routing rules. A good way to store them is in a map, for example: `map[string]map[string]func(*response.Writer, *request.Request)`. The first map key would be the HTTP method (e.g., "GET"), and the second would be the path (e.g., "/").

2.  **Create a `ServeHTTP` method for the `Mux`:**
    *   This method will be called for each incoming request. It will look up the appropriate handler in the routing map based on the request's method and path and then call it.
    *   If no handler is found, it should default to a "404 Not Found" response.

3.  **Integrate the `Mux` into your server:**
    *   In `cmd/httpserver/main.go`, instead of passing an anonymous function to `server.Serve`, you'll create an instance of your new `Mux`, register your routes on it, and then pass its `ServeHTTP` method to the server.

## Step 2: Advanced Routing

Once you have a basic mux, you can add more advanced features.

**Goal:** Support URL parameters (e.g., `/users/{id}`) and query strings.

**How-to:**

1.  **URL Parameter Parsing:**
    *   Modify your `Mux` to handle paths with placeholders (e.g., `/users/{id}`). When a request comes in, you'll need to parse the URL, extract the value (e.g., the `id`), and make it available to your handler. You could add a `Params` map to the `request.Request` struct.

2.  **Query String Parsing:**
    *   In your `request.Request` parsing logic, after you've identified the `RequestTarget`, parse it to separate the path from the query string (the part after `?`). You can then parse the query string into a map of key-value pairs and store it in the `Request` struct.

## Step 3: Middleware

Middleware allows you to chain functions to process a request before it reaches its final handler. This is great for logging, authentication, compression, etc.

**Goal:** Create a middleware system that can wrap your handlers.

**How-to:**

1.  **Define a `Middleware` type:**
    *   A middleware is essentially a function that takes a handler and returns a new handler. The new handler will contain the middleware logic and will eventually call the original handler.

2.  **Create a middleware chain:**
    *   Allow multiple middlewares to be chained together. The request will pass through each middleware in order until it reaches the final handler.

3.  **Example Middleware:**
    *   **Logger:** A simple middleware that logs the request method, path, and how long it took to process.
    *   **Authentication:** A middleware that checks for an `Authorization` header and rejects the request if it's not valid.

## Step 4: Enhancing the `response.Writer` (Done)

The `response.Writer` has been enhanced with helper functions for common response types.

**Features:**

*   **`JSON()` method:** Easily send JSON responses.
*   **`Respond200`, `Respond400`, `Respond404`, `Respond500`:** Helper functions for standard HTTP responses.

## Step 5: Static File Server (Done)

A static file server has been implemented to serve directories of static files.

**Features:**

*   Serves files from a specified directory.
*   Handles content types based on file extension.
*   Defaults to serving `index.html` for directory requests.


## Step 6: Configuration and Graceful Shutdown

Improve the server's startup and shutdown procedures.

**Goal:** Make the server more robust and production-ready.

**How-to:**

1.  **Configuration Struct:**
    *   Instead of hardcoding values, create a `Config` struct that holds settings like the port, read/write timeouts, etc. This can be populated from command-line flags or a configuration file.

2.  **Improved Graceful Shutdown:**
    *   Your current graceful shutdown is a good start. You can improve it by keeping track of active connections and waiting for them to finish before shutting down the server.

By following these steps, you'll gradually build a powerful and flexible HTTP server that will give you a much deeper understanding of how web frameworks operate.
