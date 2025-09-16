# Code Explanation: `internal/request/request.go`

This file is the heart of the server's ability to understand clients. It defines a state machine that reads a raw stream of bytes from the network and turns it into a structured `Request` object that the rest of the application can easily use.

---

### `parserState` and `const`

```go
type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)
```
- We start by defining a custom type `parserState`, which is an alias for `string`. 
- The `const` block then defines the possible states our parser can be in. Using a custom type and constants like this (an "enum" pattern) makes the code much safer and more readable than using plain strings everywhere. It prevents typos and makes the logic of the state machine explicit.

---

### `Request` and `RequestLine` Structs

```go
type RequestLine struct { ... }

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	PathParams  map[string]string
	Query       url.Values

	state parserState
}
```
- These structs hold the parsed data. 
- `RequestLine` holds the three parts of the first line of an HTTP request.
- `Request` is the main container. It holds the `RequestLine`, a pointer to the `Headers` object, the request `Body` as a string, and maps for `PathParams` (from the router) and `Query` parameters.
- `url.Values` is a type from Go's standard library, defined as `map[string][]string`. We use it to correctly handle query strings where a key can appear multiple times (e.g., `?tag=a&tag=b`).
- The `state` field is un-exported and holds the current `parserState` for this request object as it moves through the parsing process.

---

### `func parseRequestLine(...)`

This function parses the first line of the request (e.g., `GET /foo?bar=baz HTTP/1.1`).

```go
func parseRequestLine(b []byte) (*RequestLine, int, url.Values, error) {
```
- Note the four return values: the parsed `*RequestLine`, the number of bytes read (`int`), the parsed `url.Values` from the query string, and a potential `error`.

```go
	idx := bytes.Index(b, SEPARATOR) // SEPARATOR is `\r\n`
	if idx == -1 {
		return nil, 0, nil, nil
	}
```
- It first looks for the end-of-line marker (`\r\n`). If it's not found, it means we don't have a full line of data yet. It returns `0` for bytes read, signaling to the caller that it needs to wait for more data from the network.

```go
	path, rawQuery, _ := strings.Cut(requestTarget, "?")
```
- `strings.Cut` is a convenient Go function that splits a string on the first instance of a separator. It returns the part before the separator, the part after, and a boolean indicating if the separator was found. Here, it cleanly separates the URL path from the query string.

```go
	query, err := url.ParseQuery(rawQuery)
```
- `url.ParseQuery` is a powerful standard library function that parses a raw query string (`foo=bar&baz=1`) into the `url.Values` map.

---

### `func (r *Request) parse(data []byte) (int, error)`

This method is the core of the state machine. It takes a chunk of bytes and attempts to advance the parsing state.

```go
func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		// ...
		switch r.state {
```
- It uses an infinite `for` loop that will run as long as there is data to process. The `switch r.state` block determines what logic to run based on the current state.

- **`case StateInit:`**
    - It calls `parseRequestLine`. If `n` (bytes read) is `0`, it means the line was incomplete, so it breaks the loop to wait for more data. 
    - On success, it populates `r.RequestLine` and `r.Query`, and transitions the state: `r.state = StateHeaders`.

- **`case StateHeaders:`**
    - It calls the `r.Headers.Parse` method we defined in the `headers` package.
    - If that method returns `done = true` (meaning it found the `\r\n\r\n` that ends the header block), it checks if a body is expected using `r.hasBody()`.
    - It then transitions to the next state accordingly: `StateBody` or `StateDone`.

- **`case StateBody:`**
    - It reads the `Content-Length` header.
    - It then reads *only* the number of bytes it needs from the `data` buffer to complete the body. The `min()` helper function is crucial here to prevent reading past the end of the body if the `data` buffer contains part of a subsequent request.
    - Once `len(r.Body)` equals the `Content-Length`, it transitions: `r.state = StateDone`.

- **`case StateDone:`**
    - This state simply breaks the loop, as all parsing is complete.

---

### `func RequestFromReader(reader io.Reader) (*Request, error)`

This is the high-level function that orchestrates the entire parsing process. It's designed to work with network connections where data arrives in unpredictable chunks.

```go
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
```
- It creates a new `Request` and a buffer (`buf`) to hold data read from the network.
- The `for !request.done()` loop is the main engine. It will continue reading and parsing until the state machine transitions to `StateDone` or `StateError`.

```go
		n, err := reader.Read(buf[bufLen:])
```
- This is the key line for network reading. `reader.Read` attempts to fill the *remaining space* in the buffer (`buf[bufLen:]`). It blocks until some data is available. `n` is the number of bytes actually read.

```go
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
```
- It updates the buffer length and then calls our state machine (`request.parse`) on the *entire contents* of the buffer so far.

```go
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}
```
- After `parse` has consumed `readN` bytes, this `copy` operation shifts the remaining, un-parsed bytes to the beginning of the buffer. This is essential for handling cases where a single read from the network contains more than one request or an incomplete one.
- It then updates `bufLen` to reflect the new amount of data in the buffer, and the loop continues.

```
