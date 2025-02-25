package transport

import (
	"context"
	"github.com/eason-lee/l-rpc/protocol"
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

func (c *Client) Send(ctx context.Context, message *protocol.Message) (*protocol.Message, error) {
	// 获取连接
	trans, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	defer c.pool.Put(trans)

	// 编码消息
	codec := protocol.NewDefaultCodec()
	data, err := codec.Encode(message)
	if err != nil {
		return nil, err
	}

	// 发送并接收响应
	respData, err := trans.Send(data)
	if err != nil {
		return nil, err
	}

	// 解码响应
	return codec.Decode(respData)
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