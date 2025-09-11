package main

import (
	"log"
	"os"
	"os/signal"
	"ray8118/httpfromtcp/internal/server"
	"syscall"
)

const port = 42069

func main() {
	s, err := server.Serve(port)
	if err != nil {
		log.Fatal("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")

}
