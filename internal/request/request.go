package request

import (
	"bytes"
	"fmt"
	"io"
	"ray8118/httpfromtcp/internal/headers"
	"strconv"
)

// parserState represents the current state of our HTTP request parser.
// We move from one state to the next as we read from the data stream.
type parserState string

const (
	StateInit    parserState = "init"    // The initial state, expecting the request line (e.g., "GET / HTTP/1.1").
	StateHeaders parserState = "headers" // The state after the request line, expecting headers.
	StateBody    parserState = "body"    // The state after headers, expecting the request body.
	StateDone    parserState = "done"    // The final state, indicating the request is fully parsed.
	StateError   parserState = "error"   // A state representing a parsing error.
)

// RequestLine holds the parsed components of the first line of an HTTP request.
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// Request represents a parsed HTTP request.
type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	PathParams  map[string]string // Populated by the Mux with dynamic URL parameters.

	// state is the internal state of the parser for this request.
	state parserState
}

// getInt is a helper to safely get an integer value from headers.
func getInt(headers headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// newRequest creates and initializes a new Request object.
func newRequest() *Request {
	return &Request{
		state:      StateInit,
		Headers:    headers.NewHeaders(),
		Body:       "",
		PathParams: make(map[string]string),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

// parseRequestLine parses the first line of an HTTP request.
func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		// We don't have a full line yet, need more data.
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorMalformedRequestLine
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

// hasBody checks if the request is expected to have a body based on its headers.
func (r *Request) hasBody() bool {
	length := getInt(*r.Headers, "content-length", 0)
	return length > 0
}

// parse is the core state machine for parsing an HTTP request from a byte slice.
// It processes the data and transitions the request's state. It returns the number of bytes consumed.
func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer: // This label allows us to break out of the loop from inside the switch.
	for {
		currentData := data[read:]

		// If there's no more data to process in this chunk, break.
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			// Try to parse the request line.
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			// If n is 0, it means we don't have a full line yet. We need more data.
			if n == 0 {
				break outer
			}
			// Success! Update the request and move to the next state.
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			// Try to parse the headers.
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			// If n is 0, we need more data.
			if n == 0 {
				break outer
			}
			read += n

			// If `done` is true, we've hit the `\r\n\r\n` that marks the end of headers.
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			// Read the body based on the Content-Length header.
			length := getInt(*r.Headers, "content-length", 0)
			if length == 0 {
				// This parser doesn't support chunked encoding, so we assume Content-Length is present.
				panic("chunked not implemented")
			}

			// Read as much of the body as we have in the current data chunk.
			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			// If we've read the entire body, we're done.
			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			// Nothing more to do, break the loop.
			break outer
		default:
			panic("unhandled parser state")
		}
	}
	return read, nil
}

// done returns true if the request has been fully parsed or is in an error state.
func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

// RequestFromReader reads data from an io.Reader and parses it into a Request object.
// It handles reading from the network stream in chunks.
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// A buffer to hold the data read from the network.
	buf := make([]byte, 1024)
	bufLen := 0

	// Loop until the parser signals that it's done.
	for !request.done() {
		// Read new data from the reader, appending it to any existing data in the buffer.
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			// EOF is an expected error when the client closes the connection.
			if err == io.EOF && request.state != StateDone {
				return nil, fmt.Errorf("connection closed unexpectedly")
			}
			return nil, err
		}

		bufLen += n
		// Call the state machine to parse the data we have so far.
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		// Shift the un-parsed data to the beginning of the buffer.
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}