# HTTP Server from TCP

This project is a simple HTTP server built from scratch using raw TCP sockets in Go. It's a hands-on exercise to understand the fundamentals of how HTTP works under the hood.

## Inspiration

This project was inspired by a video from ThePrimeagen on YouTube, which is based on a course from [boot.dev](https://boot.dev). I followed the instructions and tried to build the server myself.

## How to Run

To run the HTTP server, you can use the following command:

```bash
go run cmd/httpserver/main.go 8080
```

This will start the server on port 8080. You can then send requests to it using a tool like `curl`:

```bash
curl http://localhost:8080
```
