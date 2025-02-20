package transport

import (
	"net"
	"time"
)

type Client struct {
	pool    *Pool
	network string
	address string
}

func NewClient(network, addr string) (*Client, error) {
	factory := func() (*TCPTransport, error) {
		conn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		return NewTCPTransport(conn), nil
	}

	pool, err := NewPool(PoolConfig{
		MaxIdle:     5,                 // 最大空闲连接数
		MaxActive:   20,                // 最大活跃连接数
		IdleTimeout: 30 * time.Second,  // 空闲超时时间
		Factory:     factory,
	})

	if err != nil {
		return nil, err
	}

	return &Client{
		pool:    pool,
		network: network,
		address: addr,
	}, nil
}

func (c *Client) Send(data []byte) error {
	conn, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(conn)

	return conn.Send(data)
}

func (c *Client) Receive() ([]byte, error) {
	conn, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	defer c.pool.Put(conn)

	return conn.Receive()
}

func (c *Client) Close() error {
	return c.pool.Close()
}