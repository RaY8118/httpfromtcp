# Code Explanation: `internal/mux/mux.go`

This file implements the HTTP request router, often called a "mux" or "multiplexer". Its primary responsibility is to match an incoming request to a specific handler function based on the request's URL path and HTTP method. It also handles middleware and extracts dynamic parameters from URLs.

---

### `HandlerFunc` and `Middleware` Types

```go
type HandlerFunc func(w *response.Writer, r *request.Request)
type Middleware func(HandlerFunc) HandlerFunc
```
- These lines define two **function types**. In Go, you can create new types based on function signatures. This is a powerful feature that makes code more readable and allows you to treat functions as first-class citizens (passing them as arguments, returning them from other functions, etc.).
- `HandlerFunc` defines the standard signature for all of our HTTP handlers.
- `Middleware` defines the signature for a middleware function: it's a function that takes one `HandlerFunc` and returns another `HandlerFunc`.

---

### `route` and `Mux` Structs

```go
type route struct {
	method  string
	handler HandlerFunc
	parts   []string
}

type Mux struct {
	routes []*route
}
```
- The `route` struct holds all the information for a single registered path: the HTTP method, the handler function to execute, and the path itself, broken into a slice of its component parts (e.g., `/users/{id}` becomes `["users", "{id}"]`).
- The `Mux` struct is the main router object. It contains a slice of pointers to `route` structs (`[]*route`).
- **Why a slice instead of a map?** A `map[string]HandlerFunc` would be faster for exact matches (like `/users`), but it wouldn't allow for dynamic matching of paths with parameters (like `/users/{id}`). A slice lets us iterate through all registered routes and perform more complex matching logic for each one.

---

### `func (m *Mux) HandleFunc(...)`

This method registers a new route and handler with the mux.

```go
func (m *Mux) HandleFunc(method, path string, handler HandlerFunc) {
	newRoute := &route{
		// ...
		parts: strings.Split(strings.Trim(path, "/"), "/"),
	}
	m.routes = append(m.routes, newRoute)
}
```
- It takes a method, a path string, and a handler function.
- It creates a new `route` struct, splitting the path into its parts.
- `append(m.routes, newRoute)` adds this new route to the mux's slice of routes.

---

### `func (m *Mux) ServeHTTP(...)`

This is the most important method in the file. It implements the core routing logic. By having this method, `Mux` satisfies the `httpfromtcp.Handler` interface, allowing it to be used by our server.

1.  **Path Splitting**: It first splits the incoming request's URL path into parts, just like it does for registered routes.
2.  **Route Iteration**: It then enters a `for` loop, iterating over every `route` that has been registered.
3.  **Matching Logic**:
    - It first checks if the `method` matches. If not, it `continue`s to the next route.
    - It then checks if the number of path `parts` match. If not, it's impossible for this route to match, so it continues.
    - It then loops through the path parts, comparing the registered part with the request's part.
    - **Parameter Extraction**: If a registered part looks like a parameter (e.g., starts with `{` and ends with `}`), it doesn't compare the text. Instead, it extracts the parameter name (e.g., `id`) and stores the corresponding value from the request's path in a temporary `params` map.
    - **Static Matching**: If the part is not a parameter, it must match the request's path part exactly. If it doesn't, the `match` flag is set to `false` and the inner loop breaks.
4.  **Execution**: 
    - If, after checking all the parts, the `match` flag is still `true`, we have found our handler.
    - `r.PathParams = params`: The extracted parameters are added to the `Request` object so the handler can access them.
    - `route.handler(w, r)`: The registered handler function is executed.
    - `return`: This is crucial. It stops the `ServeHTTP` function immediately, preventing it from matching any other routes.
5.  **404 Not Found**: If the `for` loop finishes without finding any matching routes (i.e., without hitting a `return`), it means no handler was found. The code then proceeds to write a standard `404 Not Found` response.

---

### Middleware Functions

#### `func Chain(h HandlerFunc, mws ...Middleware) HandlerFunc`
- This function takes a final handler and a variable number of middlewares (`mws ...Middleware`). The `...` syntax means it can accept zero or more `Middleware` arguments, which will be available inside the function as a slice named `mws`.
- It then loops through the middlewares **in reverse**.
- **Why reverse?** Consider `Chain(myHandler, log, auth)`. We want the request to flow through `log`, then `auth`, then `myHandler`. To achieve this, we have to wrap them from the inside out:
    1. Start with `handler = myHandler`
    2. First iteration (auth): `handler = auth(myHandler)`
    3. Second iteration (log): `handler = log(auth(myHandler))`
- The final result is a single `HandlerFunc` that contains the entire chain.

#### `func LoggingMiddleware(next HandlerFunc) HandlerFunc`
- This is an actual middleware implementation. It conforms to the `Middleware` function type.
- It takes the *next* handler in the chain as an argument (`next`).
- It returns a *new* handler function. This new function is a **closure**—it has access to the `next` variable from its parent function.
- **The Logic**: When this returned handler is called, it first records the start time, then calls `next(w, r)` to pass control to the next middleware or the final handler. When `next` eventually returns, the code after it executes, calculating the duration and printing the log message. This "before and after" execution is the essence of middleware.
