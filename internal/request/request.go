package request

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"ray8118/httpfromtcp/internal/headers"
	"strconv"
	"strings"
)

// parserState represents the current state of our HTTP request parser.
type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
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
	PathParams  map[string]string
	Query       url.Values // Correct type for query parameters

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
		Query:      make(url.Values), // Correct initialization
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

// parseRequestLine parses the first line of an HTTP request.
// It now returns the parsed query parameters as url.Values.
func parseRequestLine(b []byte) (*RequestLine, int, url.Values, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, nil, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, nil, ErrorMalformedRequestLine
	}

	// Separate path and query string
	requestTarget := string(parts[1])
	path, rawQuery, _ := strings.Cut(requestTarget, "?")

	// Parse the raw query string
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		query = make(url.Values) // On error, return empty values
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: path, // Set RequestTarget to only the path part
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, query, nil
}

// hasBody checks if the request is expected to have a body.
func (r *Request) hasBody() bool {
	length := getInt(*r.Headers, "content-length", 0)
	return length > 0
}

// parse is the core state machine for parsing an HTTP request.
func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			// Capture all return values from parseRequestLine
			rl, n, q, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			// Assign the parsed values
			r.RequestLine = *rl
			r.Query = q // Assign the parsed query
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(*r.Headers, "content-length", 0)
			if length == 0 {
				panic("chunked not implemented")
			}
			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer
		default:
			panic("unhandled parser state")
		}
	}
	return read, nil
}

// done returns true if the request has been fully parsed.
func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

// RequestFromReader reads from an io.Reader and parses it into a Request.
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if err == io.EOF && request.state != StateDone {
				return nil, fmt.Errorf("connection closed unexpectedly")
			}
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
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