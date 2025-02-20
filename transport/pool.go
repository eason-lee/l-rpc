package transport

import (
	"errors"
	"sync"
	"time"
)

type Pool struct {
	mu      sync.Mutex
	conns   chan *TCPTransport
	factory func() (*TCPTransport, error)
	closed  bool

	// 连接池配置
	maxIdle     int           // 最大空闲连接数
	maxActive   int           // 最大活跃连接数
	idleTimeout time.Duration // 空闲超时时间
}

type PoolConfig struct {
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
	Factory     func() (*TCPTransport, error)
}

func NewPool(config PoolConfig) (*Pool, error) {
	if config.MaxIdle <= 0 || config.MaxActive <= 0 {
		return nil, errors.New("invalid pool config")
	}
	if config.Factory == nil {
		return nil, errors.New("factory func is required")
	}

	return &Pool{
		conns:       make(chan *TCPTransport, config.MaxIdle),
		factory:     config.Factory,
		maxIdle:     config.MaxIdle,
		maxActive:   config.MaxActive,
		idleTimeout: config.IdleTimeout,
	}, nil
}

func (p *Pool) Get() (*TCPTransport, error) {
	if p.closed {
		return nil, errors.New("pool is closed")
	}

	// 优先从空闲连接中获取
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		return p.factory()
	}
}

func (p *Pool) Put(conn *TCPTransport) error {
	if p.closed {
		return conn.Close()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果空闲连接达到最大值，直接关闭
	if len(p.conns) >= p.maxIdle {
		return conn.Close()
	}

	select {
	case p.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true

	// 关闭所有连接
	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}

	return nil
}