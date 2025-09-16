# Code Explanation: `httpfromtcp.go`

This file defines the public API for your library. It provides a clean, high-level interface for users to start a server, hiding the internal details of listeners, parsers, and routers. The design deliberately mimics Go's standard `net/http` package to provide a familiar experience.

---

### `Handler` Interface

```go
type Handler interface {
	ServeHTTP(w *response.Writer, r *request.Request)
}
```
- This is the central interface of the library. An `interface` in Go defines a contract: a set of methods. 
- Any type that implements all the methods of an interface is said to satisfy that interface. Crucially, this is done implicitly—there is no `implements` keyword in Go. If your type has a `ServeHTTP` method with the correct signature, it *is* a `httpfromtcp.Handler`.
- This is what allows our `Mux` struct to be used as a handler, because it has a `ServeHTTP` method.

---

### `HandlerFunc` Type and Method

```go
type HandlerFunc func(w *response.Writer, r *request.Request)

func (f HandlerFunc) ServeHTTP(w *response.Writer, r *request.Request) {
	f(w, r)
}
```
- This is a clever and common pattern in Go known as an **adapter**. It lets us use an ordinary function as a `Handler`.
- `type HandlerFunc ...`: First, we define a function type that matches the signature of `ServeHTTP`.
- `func (f HandlerFunc) ServeHTTP(...)`: Second, we attach a `ServeHTTP` method to that function type.
- **How it works:** The `ServeHTTP` method simply calls the function itself (`f(w, r)`). When you have a function `myFunc` and you wrap it like `HandlerFunc(myFunc)`, you are converting it to the `HandlerFunc` type. Since that type has a `ServeHTTP` method, it now satisfies the `Handler` interface!
- This is what allowed us to pass our middleware chain (which results in a single function) to `ListenAndServe`.

---

### `func ListenAndServe(addr string, handler Handler) error`

This is the main entry point for the entire library. It's the only function a user needs to call to start a fully functional server.

```go
func ListenAndServe(addr string, handler Handler) error {
```
- The `handler` parameter is of type `Handler` (the interface). This means you can pass any type that satisfies the interface, such as a `*Mux` or a `HandlerFunc`.

```go
	s, err := server.Serve(port, handler.ServeHTTP)
```
- This line is the bridge between the public API and the internal implementation. It calls our internal `server.Serve` function.
- Notice that it passes `handler.ServeHTTP`. It's not passing the `handler` object itself, but rather the *method* associated with it. Since `server.Serve` expects a plain function (`server.Handler`), this works perfectly. The method `ServeHTTP` has the exact signature that the `server.Handler` function type requires.

```go
	defer s.Close()
```
- This `defer` statement ensures that the server's `Close()` method is called when `ListenAndServe` is about to return. This happens after a shutdown signal is received.

```go
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
```
- This is the standard Go pattern for graceful shutdown.
- `make(chan os.Signal, 1)` creates a **channel**. A channel is a conduit for communication between goroutines. This channel is designed to carry `os.Signal` values. The `1` means it's a buffered channel with a capacity of one, which is important to prevent deadlocks in some signal handling scenarios.
- `signal.Notify` tells the Go runtime to listen for specific operating system signals (`syscall.SIGINT` is Ctrl+C, `syscall.SIGTERM` is a standard termination signal) and send them into the `sigChan` channel instead of immediately terminating the program.

```go
	<-sigChan
```
- This is a **blocking receive** from the channel. The `main` goroutine, which is running `ListenAndServe`, will pause at this line and wait indefinitely. It will only un-pause when a value is sent into the `sigChan` (i.e., when the user presses Ctrl+C).
- This single line is what keeps the server running and prevents the `main` function from exiting immediately.

```go
	log.Println("Server gracefully stopped")
	return nil
}
```
- Once a signal is received and the `<-sigChan` line unblocks, the function prints a shutdown message and returns. At this point, the `defer s.Close()` statement from earlier is executed, and the server is cleanly shut down.
