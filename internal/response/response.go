package response

import (
	"encoding/json"
	"fmt"
	"io"
	"ray8118/httpfromtcp/internal/headers"
	"ray8118/httpfromtcp/internal/request"
)

type Response struct {
}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusCreated             StatusCode = 201
	StatusNotFound            StatusCode = 404
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func Respond200(w *Writer) {
	body := []byte(`
	<html>
	<head>
		<title>200 OK</title>
	</head>
	<body>
		<h1>Success!</h1>
		<p>Your request was an absolute banger.</p>
	</body>
	</html>
	`)
	h := GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(StatusOk)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func Respond400(w *Writer) {
	body := []byte(`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
	`)
	h := GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(StatusBadRequest)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func Respond404(w *Writer) {
	body := []byte(`
<html>
  <head>
    <title>404 Not found</title>
  </head>
  <body> 
    <h1>Not found</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
	`)
	h := GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(StatusBadRequest)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func Respond500(w *Writer) {
	body := []byte(`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`)
	h := GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(StatusInternalServerError)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

// JSON marshals the provided data strcuture to JSON, sets the
// appropriate headers, and writes the response
func (w *Writer) JSON(statusCode int, data interface{}) {
	// Marshall the data into a JSON byte slice
	jsonData, err := json.Marshal(data)
	if err != nil {
		// if marshalling fails, it's a server-side problem
		// Wr send a 500 Internal Server Error
		w.WriteStatusLine(StatusInternalServerError)
		h := GetDefaultHeaders(0)
		w.WriteHeaders(*h)
	}

	// Set the status line and headers
	w.WriteStatusLine(StatusCode(statusCode))
	h := GetDefaultHeaders(len(jsonData))
	h.Set("Content-Type", "application/json")
	w.WriteHeaders(*h)
	// Write the JSON body
	w.WriteBody(jsonData)

}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOk:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusCreated:
		statusLine = []byte("HTTP/1.1 201 Created\r\n")
	case StatusNotFound:
		statusLine = []byte("HTTP/1.1 404 Not Found\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unrecognized error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err

}
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	return n, err
}
