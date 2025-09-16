# Code Explanation: `internal/response/response.go`

This file defines the tools for writing an HTTP response back to the client. Its central component is the `Writer` struct, which abstracts the process of sending the status line, headers, and body.

---

### Type and Constant Definitions

```go
type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusCreated             StatusCode = 201
	StatusNotFound            StatusCode = 404
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)
```
- `type StatusCode int` defines a new, distinct type named `StatusCode` that has the same underlying structure as an `int`. This is a common Go practice to improve type safety. It prevents you from accidentally using a generic `int` where a status code is expected.
- The `const` block defines a set of named constants for common HTTP status codes. This makes the code much more readable than using raw numbers like `200` or `404` directly in your handlers.

---

### `type Writer struct { ... }`

This is the struct that your handlers will interact with.

```go
type Writer struct {
	writer io.Writer
}
```
- It has one un-exported field, `writer`.
- The type of this field, `io.Writer`, is one of the most important concepts in Go. It is an **interface**. An interface defines a set of methods, and any type that implements those methods automatically satisfies the interface.
- The `io.Writer` interface requires only one method: `Write(p []byte) (n int, err error)`.
- By using `io.Writer`, our `response.Writer` doesn't care *where* it's writing to. It could be a network connection (like in our case), a file on disk, an in-memory buffer for testing, or even standard output. This makes the code incredibly flexible and easy to test.

---

### `func NewWriter(writer io.Writer) *Writer`

This is the constructor for our `Writer`.

```go
func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}
```
- It takes an `io.Writer` (which, in our server, will be the `net.Conn` representing the TCP connection to the client) and stores it inside the `Writer` struct.
- It returns a pointer (`*Writer`) so that the methods called on it operate on the original struct, not a copy.

---

### `func (w *Writer) WriteStatusLine(statusCode StatusCode) error`

This method writes the first line of an HTTP response (e.g., `HTTP/1.1 200 OK`).

```go
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	// ... cases
	}

	_, err := w.writer.Write(statusLine)
	return err
}
```
- It uses a `switch` statement on the `statusCode` to select the correct, full status line string.
- It then calls the `Write` method of the underlying `io.Writer` (`w.writer`) to send the byte slice over the connection.
- It returns any error that might occur during the write operation (e.g., if the client has disconnected).

---

### `func (w *Writer) WriteHeaders(h headers.Headers) error`

This method writes the block of headers.

```go
func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}
```
- It uses the `ForEach` method from our `headers` package to iterate over each header.
- `fmt.Appendf` is a useful and efficient function that formats a string and appends it to a byte slice, reallocating the slice if necessary.
- After appending all the header lines, it appends one final `\r\n` to signify the end of the header block.
- Finally, it writes the entire block of bytes to the `io.Writer`.

---

### `func (w *Writer) WriteBody(p []byte) (int, error)`

This method writes the response body.

```go
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}
```
- This is a simple wrapper around the underlying `io.Writer`'s `Write` method. It passes the body content directly to it and returns the result.

---

### `func (w *Writer) JSON(statusCode int, data interface{})`

This is the helper method we built to simplify sending JSON responses.

```go
func (w *Writer) JSON(statusCode int, data interface{}) {
```
- The `data` parameter is of type `interface{}`. This is the **empty interface** in Go. It can hold a value of *any* type, because all types implement the empty interface (which has no methods). This allows us to pass any Go struct or map to this function to be marshaled into JSON.

```go
	jsonData, err := json.Marshal(data)
	if err != nil {
		// ... error handling
	}
```
- `json.Marshal` is the standard library function that converts a Go data structure into a JSON byte slice.

```go
	w.WriteStatusLine(StatusCode(statusCode))
	h := GetDefaultHeaders(len(jsonData))
	h.Set("Content-Type", "application/json")
	w.WriteHeaders(*h)
	w.WriteBody(jsonData)
}
```
- This part orchestrates the entire response. It calls the other methods on the `Writer` in the correct order: status line, then headers, then the body. This is a great example of how helper methods can be composed to build more powerful functionality.

```
