package static

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"ray8118/httpfromtcp/internal/request"
	"ray8118/httpfromtcp/internal/response"
	"strings"
)

func Static(w *response.Writer, r *request.Request) {
	relPath := strings.TrimPrefix(r.RequestLine.RequestTarget, "/static")

	if relPath == "" || relPath == "/" {
		relPath = "index.html"
	}

	// Always ensure a leading slash before joining
	if !strings.HasPrefix(relPath, "/") {
		relPath = "/" + relPath
	}

	path := "static" + relPath

	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		path = filepath.Join(path, "index.html")
		fmt.Println("Directory requested, serving:", path)
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		response.Respond200(w)
		return
	}
	if err != nil {
		response.Respond500(w)
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	h := response.GetDefaultHeaders(int(stat.Size()))
	h.Replace("Content-Type", mimeType)
	w.WriteStatusLine(response.StatusOk)
	w.WriteHeaders(*h)

	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			w.WriteBody(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading file: %v", err)
			break
		}
	}
}
