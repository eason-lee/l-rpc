package transport

import (
    "net"
)

// Client 传输层客户端
type Client struct {
    transport Transport
}

func NewClient(network, addr string) (*Client, error) {
    conn, err := net.Dial(network, addr)
    if err != nil {
        return nil, err
    }
    return &Client{
        transport: NewTCPTransport(conn),
    }, nil
}

func (c *Client) Send(data []byte) error {
    return c.transport.Send(data)
}

func (c *Client) Receive() ([]byte, error) {
    return c.transport.Receive()
}

func (c *Client) Close() error {
    return c.transport.Close()
}