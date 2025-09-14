This document explains how the HTTP server project works, breaking it down file by file and function by function.

## Project Overview

The project is a simple HTTP server built in Go using raw TCP sockets. It's designed to demonstrate the basic principles of the HTTP protocol without relying on Go's built-in HTTP libraries. The server listens for TCP connections, parses incoming data as HTTP requests, and sends back HTTP responses.

## File by File Explanation

### `go.mod`

*   **Purpose**: This file is the standard way to manage dependencies in a Go project. It defines the module's path (`ray8118/httpfromtcp`) and lists the external libraries the project depends on. In this case, it's `github.com/stretchr/testify`, which is a popular library for writing tests.

### `messages.txt`

*   **Purpose**: This is a simple text file that acts as the content source for our HTTP server. When you make a `GET` request to the server's root (`/`), the content of this file is read, processed, and sent back as the body of the HTTP response.

### `cmd/httpserver/main.go`

*   **Purpose**: This is the main entry point for running the HTTP server.
*   **`main()`**: This function does the following:
    1.  It checks for a port number from the command-line arguments.
    2.  It starts a TCP listener on the specified port using `net.Listen`.
    3.  It enters an infinite loop to accept new TCP connections using `l.Accept()`.
    4.  For each new connection, it spawns a new goroutine (a lightweight thread) to handle the connection by calling `server.HandleConnection`. This allows the server to handle multiple clients concurrently.

### `cmd/tcplistener/main.go`

*   **Purpose**: This is a simple utility program for debugging. It listens on a TCP port and prints any data it receives to the console. This is useful for testing what a client is sending.

### `cmd/udpsender/main.go`

*   **Purpose**: This is another utility program. It sends a message provided on the command line to a specified host and port using the UDP protocol. It's not directly related to the HTTP server but could be used for other networking tests.

### `internal/headers/headers.go`

*   **Purpose**: This package handles the parsing and formatting of HTTP headers.
*   **`Headers` type**: This is a `map[string]string` that stores header key-value pairs.
*   **`String()` method**: This method converts the `Headers` map into the correct string format required by the HTTP protocol (e.g., `Key: Value\r\n`).
*   **`Parse()` function**: This function reads from a `bufio.Reader` line by line, parsing the HTTP headers until it encounters a blank line, which signifies the end of the headers section.

### `internal/request/request.go`

*   **Purpose**: This package is responsible for parsing an incoming stream of bytes and turning it into a structured `Request` object.
*   **`Request` struct**: This struct represents an HTTP request, containing the `Method` (e.g., GET), `Path` (e.g., /), `Version` (e.g., HTTP/1.1), `Headers`, and `Body`.
*   **`Parse()` function**: This is the core of the request parsing logic. It does the following:
    1.  Reads the first line, which is the "request line," and splits it into the method, path, and version.
    2.  Calls `headers.Parse` to handle the header section.
    3.  Checks for a `Content-Length` header. If present, it reads that many bytes from the reader to get the request body.
    4.  It returns a populated `Request` struct.

### `internal/response/response.go`

*   **Purpose**: This package defines the structure of an HTTP response and provides a way to format it as a string.
*   **`Response` struct**: This struct holds all the components of an HTTP response: `Version`, `StatusCode` (e.g., 200), `StatusText` (e.g., OK), `Headers`, and `Body`.
*   **`String()` method**: This method assembles the parts of the `Response` struct into a single string that conforms to the HTTP response format. This is the string that gets written back to the client over the TCP connection.

### `internal/server/server.go`

*   **Purpose**: This is where the main application logic resides.
*   **`HandleConnection()` function**: This function is the bridge between the raw TCP connection and the HTTP logic. It takes a `net.Conn` object, creates a `bufio.Reader` from it, and then calls `request.Parse` to parse the incoming HTTP request. After getting a response from `handleRequest`, it writes the response string back to the connection.
*   **`handleRequest()` function**: This function contains the routing logic. It inspects the parsed `request.Request` and decides what to do. 
    *   If the request is a `GET` to the `/` path, it reads the `messages.txt` file, prepares a 200 OK response with the file's content, and returns it.
    *   For any other request, it returns a 404 Not Found response.
