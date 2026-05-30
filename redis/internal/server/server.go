package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"

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

	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE) {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 5 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Printf("server: accept error: %v; retrying in %s", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return fmt.Errorf("server: accept connection: %w", err)
		}
		tempDelay = 0
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
