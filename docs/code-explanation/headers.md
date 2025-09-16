# Code Explanation: `internal/headers/headers.go`

This file defines the logic for handling HTTP headers. It provides a `Headers` struct that can parse headers from a raw byte stream, store them, and allow for easy access and modification.

---

### `isToken(str string) bool`

This is a helper function that is not exported (it starts with a lowercase letter, so it's only visible within the `headers` package). Its purpose is to validate if a header name is valid according to the HTTP specification.

```go
func isToken(str string) bool {
	for _, ch := range str {
```
- The `for _, ch := range str` syntax is Go's way of iterating over a string. For each character (`ch`) in the string (`str`), the loop body is executed. The `_` (underscore) is the blank identifier, used here because we don't need the index of the character, only the character itself.

```go
		found := false
		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
```
- This block checks if the character is a standard alphanumeric character.

```go
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}
```
- A `switch` statement in Go is a powerful control flow structure. Here, it's used to check if `ch` is one of the other special characters allowed in a header name.

```go
		if !found {
			return false
		}
	}
	return true
}
```
- If any character is not found in the allowed set, the function immediately returns `false`. If the loop completes without finding any invalid characters, it returns `true`.

---

### `var rn = []byte("\r\n")`

This is a package-level variable. It's a byte slice (`[]byte`) containing the Carriage Return and Line Feed characters, which are used to mark the end of a line in HTTP.

---

### `parseHeader(fieldLine []byte) (string, string, error)`

This helper function parses a single header line (like `Content-Type: application/json`).

```go
func parseHeader(fieldLine []byte) (string, string, error) {
```
- The function signature shows it takes a byte slice and returns three values: two strings (for the header name and value) and an `error`. Go functions can return multiple values, which is often used for returning a result and an error status.

```go
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
```
- `bytes.SplitN` splits the byte slice `fieldLine` by the first colon (`:`). The `2` means it will split into at most two parts. This correctly handles cases where the header value itself might contain a colon.

```go
	value := bytes.TrimSpace(parts[1])
```
- `bytes.TrimSpace` removes any leading or trailing whitespace from the header value, which is required by the HTTP spec.

---

### `type Headers struct { ... }`

This defines the main `Headers` type.

```go

type Headers struct {
	headers map[string]string
}
```
- It's a `struct`, which is Go's way of defining a collection of fields. This struct has a single, un-exported field `headers`.
- The field type is `map[string]string`. A map is Go's implementation of a hash table, used here to store header names (as keys) and their corresponding values.

---

### `func NewHeaders() *Headers`

This is a constructor function for the `Headers` struct.

```go
func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}
```
- It returns a `*Headers`, which is a **pointer** to a new `Headers` instance. Using pointers is important because it allows methods to modify the original struct. If it returned the struct directly, methods would operate on a copy.
- `map[string]string{}` is how you create an empty map.

---

### `func (h *Headers) Get(name string) (string, bool)`

This is a **method** on the `Headers` struct. It retrieves a header value by its name.

```go
func (h *Headers) Get(name string) (string, bool) {
```
- `(h *Headers)` is the **receiver**. It means this method is attached to the `Headers` type. Because it's `*Headers` (a pointer), this method can modify the `Headers` struct (though `Get` doesn't need to).

```go
	str, ok := h.headers[strings.ToLower(name)]
	return str, ok
}
```
- `strings.ToLower(name)` ensures that header lookups are case-insensitive, as required by the HTTP spec.
- The `str, ok := ...` is a common Go idiom for accessing a map. If the key (`name`) exists in the map, `str` will be the value and `ok` will be `true`. If not, `str` will be the zero value for a string (i.e., `""`) and `ok` will be `false`.

---

### `Set`, `Replace`, `Delete` Methods

These methods all have a pointer receiver `(h *Headers)` because they **modify** the internal `headers` map.

- **`Replace`**: Simply overwrites any existing value for the header.
- **`Delete`**: Uses the built-in `delete()` function to remove a key from the map.
- **`Set`**: This one is interesting. If the header already exists, it appends the new value with a comma. This is how HTTP allows for a single header key to have multiple values (e.g., `Cache-Control: no-cache, no-store`).

---

### `func (h *Headers) ForEach(cb func(n, v string))`

This method iterates over all the headers and calls a callback function for each one.

```go
func (h *Headers) ForEach(cb func(n, v string)) {
```
- The parameter `cb` is a **function**. In Go, functions are first-class citizens, meaning they can be passed as arguments to other functions. This is a powerful feature for creating flexible and extensible APIs.

---

### `func (h Headers) Parse(data []byte) (int, bool, error)`

This is the core parsing logic. It reads a block of bytes and extracts all the header lines.

```go
func (h Headers) Parse(data []byte) (int, bool, error) {
```
- Notice the receiver is `(h Headers)`, not `*Headers`. This is a subtle point. The method *does* modify the headers map, which might seem to contradict the rule about using pointer receivers for modification. However, a `map` in Go is a **reference type**. This means that even when the `Headers` struct is copied, the `headers` field inside it still points to the *same underlying map data*. Therefore, changes made here will persist.

```go
	for {
```
- `for {}` is an infinite loop in Go, equivalent to `while(true)` in other languages. The loop is broken internally with `break` statements.

```go
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}
```
- This looks for the next `\r\n` (end of line) in the data. If it can't find one, it means we have an incomplete line, so we break the loop and wait for more data.

```go
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
```
- If `idx` is `0`, it means we've found an empty line (`\r\n` at the very start of the remaining data). An empty line signifies the end of the entire header block. We set `done = true` and break.

```go
		name, value, err := parseHeader(data[read : read+idx])
```
- This uses our `parseHeader` helper on the slice of data that contains just one header line.

```go
		h.Set(name, value)
```
- It adds the parsed header to our map using the `Set` method.

```go
	}
	return read, done, nil
}
```
- Finally, it returns the total number of bytes read, whether it finished parsing the whole block (`done`), and any error that occurred.

```
