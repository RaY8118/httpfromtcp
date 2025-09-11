package server

import (
	"fmt"
	"io"
	"net"
	"ray8118/httpfromtcp/internal/response"
)

type Server struct {
	closed   bool
	listener net.Listener
}

func runConnection(_s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, headers)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.closed {
				return
			}
			return
		}
		go runConnection(s, conn)
	}
}

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{closed: false, listener: listener}
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() {
	s.closed = true
	s.listener.Close()
}
