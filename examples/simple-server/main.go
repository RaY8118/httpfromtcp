package main

import (
	"log"

	"ray8118/httpfromtcp"
	"ray8118/httpfromtcp/internal/mux"
	"ray8118/httpfromtcp/internal/static"
)

const addr = ":42069"

func main() {
	// Create a new mux from our library
	m := mux.NewMux()

	m.HandleFunc("GET", "/", handleRoot)
	m.HandleFunc("GET", "/yourproblem", handleYourProblem)
	m.HandleFunc("GET", "/myproblem", handleMyProblem)
	m.HandleFunc("GET", "/video", handleVideo)
	m.HandleFunc("GET", "hello/{name}", handleHelloUser)
	m.HandleFunc("POST", "/messages", handleCreateMessage)
	m.HandleFunc("GET", "/query-test", handleQueryTest)
	m.HandleFunc("GET", "/user", handlerUserJSON)
	m.HandleFunc("POST", "/user", handleCreateUser)
	m.HandleFunc("GET", "/static", static.Static)

	m.HandleFunc("GET", "/httpbin/get", handleHttpbin)
	m.HandleFunc("GET", "/httpbin/ip", handleHttpbin)
	m.HandleFunc("GET", "/httpbin/user-agent", handleHttpbin)

	log.Printf("Starting server on %s", addr)

	// Chain the middleware to the mux's ServeHTTP method.
	chainedHandler := mux.Chain(m.ServeHTTP, mux.LoggingMiddleware)

	// Convert the resulting HandlerFunc back into a Handler that ListenAndServe can accept.
	err := httpfromtcp.ListenAndServe(addr, httpfromtcp.HandlerFunc(chainedHandler))
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
