package transport

import (
    "net"
)

// Server 传输层服务端
type Server struct {
    listener net.Listener
    handler  func(Transport)
}

func NewServer(addr string) (*Server, error) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }
    return &Server{listener: listener}, nil
}

// Accept 接受新的连接
func (s *Server) Accept(handler func(Transport)) error {
    s.handler = handler
    for {
        conn, err := s.listener.Accept()
        if err != nil {
            return err
        }
        // 为每个连接创建一个 goroutine
        go s.handleConn(conn)
    }
}

func (s *Server) handleConn(conn net.Conn) {
    transport := NewTCPTransport(conn)
    defer transport.Close()
    s.handler(transport)
}