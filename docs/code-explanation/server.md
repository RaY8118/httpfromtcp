# Code Explanation: `internal/server/server.go`

This file is the engine room of the web server. It handles the raw networking: listening for TCP connections, accepting them, and then spawning concurrent processes to handle each one. It bridges the gap between the low-level network and the high-level HTTP parsing and routing.

---

### `Handler` and `Server` Types

```go
type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed  bool
	handler Handler
}
```
- `Handler` is a function type that defines the signature for the main handler that the server will call for every valid request. In our project, this will be the `ServeHTTP` method of our `Mux`.
- `Server` is a struct that holds the state of the server. It contains the `handler` function it needs to call and a `closed` boolean flag used for graceful shutdown.

---

### `func runConnection(s *Server, conn io.ReadWriteCloser)`

This function's logic is executed for every single client connection.

```go
func runConnection(s *Server, conn io.ReadWriteCloser) {
```
- The `conn` parameter is of type `io.ReadWriteCloser`. This is a Go interface that combines three other interfaces: `io.Reader`, `io.Writer`, and `io.Closer`. A network connection (`net.Conn`) satisfies all three, so it can be passed into this function.

```go
	defer conn.Close()
```
- `defer` is a powerful keyword in Go. It schedules the `conn.Close()` function call to be executed right before `runConnection` returns. This is a guarantee. No matter what happens in the function—whether it finishes successfully or returns early due to an error—the connection will be closed. This is the standard, idiomatic way to manage resources like network connections or files in Go.

```go
	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
```
- Here, the single `conn` object is used in two ways:
    1. It's passed to `response.NewWriter`, which uses it as an `io.Writer` to send data *to* the client.
    2. It's passed to `request.RequestFromReader`, which uses it as an `io.Reader` to receive data *from* the client.

```go
	if err != nil {
		// ... send 400 Bad Request
		return
	}
```
- If `RequestFromReader` returns an error (e.g., the client sent a malformed request), the function writes a `400 Bad Request` status back to the client and `return`s. The deferred `conn.Close()` is executed at this point.

```go
	s.handler(responseWriter, r)
}
```
- If parsing is successful, it calls the main handler that was configured for the server, passing it the `responseWriter` and the parsed `request`.

---

### `func runServer(s *Server, listener net.Listener)`

This function is the main loop that waits for new connections.

```go
func runServer(s *Server, listener net.Listener) {
	for {
```
- `for {}` creates an infinite loop, causing the server to listen for connections forever until the program is terminated.

```go
		conn, err := listener.Accept()
```
- `listener.Accept()` is a **blocking call**. The code execution will pause here and wait until a new client connects to the server. When a client connects, this function returns a `net.Conn` object (representing the connection) and an error.

```go
		if err != nil {
			if s.closed { // ... }
			return
		}
```
- If an error occurs, it first checks if the server was intentionally closed (`s.closed`). If so, this error is expected, and the function just returns. Otherwise, it logs the error.

```go
		go runConnection(s, conn)
	}
}
```
- This is the core of the server's concurrency model. The `go` keyword starts a new **goroutine**. 
- A goroutine is a lightweight thread of execution managed by the Go runtime. You can think of it as a function that runs in the background, independently of the main flow.
- By calling `go runConnection(...)`, the server immediately starts handling the new connection in the background. Crucially, it does **not** wait for `runConnection` to finish. The `for` loop immediately continues to the next iteration and calls `listener.Accept()` again, ready to handle the next client.
- This is how a Go server can handle thousands of concurrent connections efficiently without complex threading logic.

---

### `func Serve(port uint16, handler Handler) (*Server, error)`

This is the main entry point for this package.

```go
func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
```
- `net.Listen` is the standard library function that opens a network port and returns a `net.Listener`.

```go
	server := &Server{closed: false, handler: handler}

	go runServer(server, listener)

	return server, nil
}
```
- It creates the `Server` struct.
- It then starts the main accept loop (`runServer`) in its own goroutine. This is done so that the `Serve` function itself doesn't block and can return immediately.
- It returns the `*Server` object to the caller. This allows the caller to interact with the server after it has started, for example, by calling its `Close()` method.
