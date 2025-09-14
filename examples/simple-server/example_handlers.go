package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"ray8118/httpfromtcp/internal/headers"
	"ray8118/httpfromtcp/internal/mux"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
)

func respond400(w *response.Writer) {
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
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(response.StatusBadRequest)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func respond500(w *response.Writer) {
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
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(response.StatusInternalServerError)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func respond200(w *response.Writer) {
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
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-type", "text/html")
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func handleRoot(w *response.Writer, r *request.Request) {
	respond200(w)
}

func handleYourProblem(w *response.Writer, r *request.Request) {
	respond400(w)
}

func handleMyProblem(w *response.Writer, r *request.Request) {
	respond500(w)
}

func handleVideo(w *response.Writer, r *request.Request) {
	// 1. Open the file. This doesn't load it into memory.
	f, err := os.Open("assets/vim.mp4")
	if err != nil {
		respond500(w)
		return
	}
	defer f.Close()

	// 2. Get file info to find its size for the Content-Length header.
	stat, err := f.Stat()
	if err != nil {
		respond500(w)
		return
	}
	fileSize := stat.Size()

	// 3. Write the status and headers.
	h := response.GetDefaultHeaders(int(fileSize)) // Note: This might need adjustment if fileSize is very large
	h.Replace("content-type", "video/mp4")
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)

	// 4. Stream the file in chunks.
	// Create a buffer to hold parts of the file. 32KB is a reasonable size.
	buf := make([]byte, 32*1024)
	for {
		// Read a chunk from the file into the buffer.
		n, err := f.Read(buf)
		if n > 0 {
			// Write the chunk we just read to the response.
			w.WriteBody(buf[:n])
		}

		// If we're at the end of the file, `err` will be `io.EOF`.
		if err == io.EOF {
			break
		}
		// If there's any other error, stop.
		if err != nil {
			log.Printf("Error reading from file: %v", err)
			break
		}
	}
}

func handleHelloUser(w *response.Writer, r *request.Request) {
	name, ok := r.PathParams["name"]
	if !ok {
		name = "stranger"
	}

	body := []byte("Hello " + name)
	h := response.GetDefaultHeaders(len(body))
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func handleCreateMessage(w *response.Writer, r *request.Request) {
	// For a POST request, we read the body
	message := r.Body

	log.Printf("Received new message: %s", message)

	// Response with a 201 Created status
	body := []byte("Message created successfully: " + message)
	h := response.GetDefaultHeaders(len(body))
	w.WriteStatusLine(response.StatusCreated)
	w.WriteHeaders(*h)
	w.WriteBody(body)
}

func handleHttpbin(w *response.Writer, r *request.Request) {
	target := r.RequestLine.RequestTarget
	res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
	if err != nil {
		respond500(w)
		return
	}

	h := response.GetDefaultHeaders(0)
	h.Delete("Content-length")
	h.Set("transfer-encoding", "chunked")
	h.Replace("content-type", "text/plain")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)

	fullBody := []byte{}
	for {
		data := make([]byte, 32)
		n, err := res.Body.Read(data)
		if err != nil {
			break
		}
		fullBody = append(fullBody, data[:n]...)
		w.WriteBody(fmt.Appendf(nil, "%x\r\n", n))
		w.WriteBody(data[:n])
		w.WriteBody([]byte("\r\n"))
	}
	w.WriteBody([]byte("0\r\n"))
	tailers := headers.NewHeaders()
	out := sha256.Sum256(fullBody)
	tailers.Set("X-Content-SHA256", hex.EncodeToString(out[:]))
	tailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	w.WriteHeaders(*tailers)
}

func registerExampleHandlers(m *mux.Mux) {
	m.HandleFunc("GET", "/", handleRoot)
	m.HandleFunc("GET", "/yourproblem", handleYourProblem)
	m.HandleFunc("GET", "/myproblem", handleMyProblem)
	m.HandleFunc("GET", "/video", handleVideo)
	m.HandleFunc("GET", "hello/{name}", handleHelloUser)
	m.HandleFunc("POST", "/messages", handleCreateMessage)

	// This is a bit of a catch-all for the httpbin proxy.
	// A more advanced router would handle this more gracefully.
	m.HandleFunc("GET", "/httpbin/get", handleHttpbin)
	m.HandleFunc("GET", "/httpbin/ip", handleHttpbin)
	m.HandleFunc("GET", "/httpbin/user-agent", handleHttpbin)
}
