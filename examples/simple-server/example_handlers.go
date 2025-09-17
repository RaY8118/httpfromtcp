package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"ray8118/httpfromtcp/internal/headers"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
)

type UserData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func handleRoot(w *response.Writer, r *request.Request) {
	response.Respond200(w)
}

func handleYourProblem(w *response.Writer, r *request.Request) {
	response.Respond400(w)
}

func handleMyProblem(w *response.Writer, r *request.Request) {
	response.Respond500(w)
}

func handleVideo(w *response.Writer, r *request.Request) {
	// 1. Open the file. This doesn't load it into memory.
	f, err := os.Open("assets/vim.mp4")
	if err != nil {
		response.Respond500(w)
		return
	}
	defer f.Close()

	// 2. Get file info to find its size for the Content-Length header.
	stat, err := f.Stat()
	if err != nil {
		response.Respond500(w)
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

func handleQueryTest(w *response.Writer, r *request.Request) {
	var body string
	body += "Query Parameters:\n"

	for key, values := range r.Query {
		for _, v := range values {
			body += fmt.Sprintf("- %s: %s\n", key, v)
		}
	}

	w.WriteStatusLine(response.StatusOk)
	h := response.GetDefaultHeaders(len(body))
	h.Replace("Content-Type", "text/plain")
	w.WriteHeaders(*h)
	w.WriteBody([]byte(body))
}

func handlerUserJSON(w *response.Writer, r *request.Request) {
	user := UserData{
		ID:   123,
		Name: "Parth",
	}
	w.JSON(200, user)
}

func handleCreateUser(w *response.Writer, r *request.Request) {
	var reqBody CreateUserRequest

	err := json.Unmarshal([]byte(r.Body), &reqBody)
	if err != nil {
		errorResponse := map[string]string{"error": "Invalid request body"}
		w.JSON(400, errorResponse)
		return
	}

	newUser := UserData{
		ID:   456,
		Name: reqBody.Name,
	}

	w.JSON(201, newUser)
}

func handleHttpbin(w *response.Writer, r *request.Request) {
	target := r.RequestLine.RequestTarget
	res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
	if err != nil {
		response.Respond500(w)
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
