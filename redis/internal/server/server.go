package server

import (
	"fmt"
	"net"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/resp"
)

type Config struct {
	Host string
	Port string
}

type Handler func(resp.Value) (resp.Value, error)

type Server struct {
	address string
	handler Handler
}

func New(c Config, handler Handler) *Server {
	host := c.Host
	port := c.Port

	if host == "" {
		host = "0.0.0.0"
	}

	if port == "" {
		port = "6379"
	}

	return &Server{
		address: net.JoinHostPort(host, port),
		handler: handler,
	}
}

func (s *Server) Run() error {
	if s.handler == nil {
		return fmt.Errorf("server: nil handler")
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("server: listen on %s: %w", s.address, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("server: accept connection: %w", err)
		}

		go s.process(conn)
	}
}

func (s *Server) process(conn net.Conn) {
	defer conn.Close()

	dec := resp.NewDecoder(conn)
	enc := resp.NewEncoder(conn)

	for {
		var req resp.Value
		if err := dec.Decode(&req); err != nil {
			return
		}

		res, err := s.handler(req)
		if err != nil {
			// Unexpected Error
			_ = enc.Encode(resp.ErrorValue("ERR internal error"))
			return

		}
		if err := enc.Encode(res); err != nil {
			return
		}
	}
}
