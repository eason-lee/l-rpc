package transport

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"
)

// Transport 定义传输层接口
type Transport interface {
	Send(data []byte) ([]byte, error)
	Receive() ([]byte, error)
	Close() error
}

const (
	// 心跳相关常量
	heartbeatInterval = 30 * time.Second
	heartbeatTimeout  = 5 * time.Second
)

// HeartbeatMessage 心跳消息类型
type HeartbeatMessage struct {
	Type    byte  // 0: ping, 1: pong
	TimeNow int64 // 发送时间戳
}
type TransportOpts struct {
	Compressor Compressor
	Encryptor  Encryptor
}

type TCPTransport struct {
	conn           net.Conn
	heartbeatStop  chan struct{}
	lastActiveTime time.Time
	mu             sync.Mutex
	compressor     Compressor
	encryptor      Encryptor
}

func NewTCPTransport(conn net.Conn, opts ...TransportOpts) *TCPTransport {
	opt :=  TransportOpts{
		Compressor: &GzipCompressor{},
		Encryptor:  NewAESEncryptor([]byte("default-secure-key-12345")),
	}
    if len(opts) > 0 {
        opt = opts[0]
    }
	t :=  &TCPTransport{
		conn:           conn,
		heartbeatStop:  make(chan struct{}),
		lastActiveTime: time.Now(),
		compressor:     opt.Compressor,
		encryptor:      opt.Encryptor,
	}
	go t.heartbeat()
	
	return t
}

func (t *TCPTransport) heartbeat() {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := t.sendHeartbeat(); err != nil {
				t.Close()
				return
			}
		case <-t.heartbeatStop:
			return
		}
	}
}

func (t *TCPTransport) sendHeartbeat() error {
	heartbeat := HeartbeatMessage{
		Type:    0, // ping
		TimeNow: time.Now().UnixNano(),
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		return err
	}

	// 设置写入超时
	t.conn.SetWriteDeadline(time.Now().Add(heartbeatTimeout))
	defer t.conn.SetWriteDeadline(time.Time{})

	// 直接调用 send 方法，避免死锁
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.send(data)
}

func (t *TCPTransport) updateLastActiveTime() {
	t.mu.Lock()
	t.lastActiveTime = time.Now()
	t.mu.Unlock()
}

// Send 发送数据
func (t *TCPTransport) Send(data []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 发送数据
	err := t.send(data)
	if err != nil {
		return nil, err
	}

	// 接收响应
	return t.receive()
}

// receive 接收原始数据
func (t *TCPTransport) receive() ([]byte, error) {
    // 先读取数据长度
    var length uint32
    if err := binary.Read(t.conn, binary.BigEndian, &length); err != nil {
        return nil, err
    }
    
    // 读取数据内容
    data := make([]byte, length)
    _, err := io.ReadFull(t.conn, data)
    if err != nil {
        return nil, err
    }
    return data, nil
}

// Receive 接收数据
func (t *TCPTransport) Receive() ([]byte, error) {
	t.updateLastActiveTime()

	data, err := t.receive()
	if err != nil {
		return nil, err
	}

	if t.encryptor != nil {
		data, err = t.encryptor.Decrypt(data)
		if err != nil {
			return nil, err
		}
	}

	if t.compressor != nil {
		data, err = t.compressor.Decompress(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// send 发送原始数据
func (t *TCPTransport) send(data []byte) error {
	// 压缩数据
	if t.compressor != nil {
		var err error
		data, err = t.compressor.Compress(data)
		if err != nil {
			return err
		}
	}

	// 加密数据
	if t.encryptor != nil {
		var err error
		data, err = t.encryptor.Encrypt(data)
		if err != nil {
			return err
		}
	}

	// 写入数据长度
	if err := binary.Write(t.conn, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	// 写入数据内容
	_, err := t.conn.Write(data)
	return err
}

func (t *TCPTransport) Close() error {
	close(t.heartbeatStop)
	return t.conn.Close()
}
