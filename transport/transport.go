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
    Send(data []byte) error
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
    Type    byte   // 0: ping, 1: pong
    TimeNow int64  // 发送时间戳
}

// TCPTransport TCP 传输层实现
type TCPTransport struct {
    conn              net.Conn
    heartbeatStop    chan struct{}
    lastActiveTime   time.Time
    mu               sync.Mutex
}

func NewTCPTransport(conn net.Conn) *TCPTransport {
    t := &TCPTransport{
        conn:           conn,
        heartbeatStop: make(chan struct{}),
        lastActiveTime: time.Now(),
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
    t.mu.Lock()
    defer t.mu.Unlock()

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

    return t.Send(data)
}

func (t *TCPTransport) updateLastActiveTime() {
    t.mu.Lock()
    t.lastActiveTime = time.Now()
    t.mu.Unlock()
}

// Send 发送数据
func (t *TCPTransport) Send(data []byte) error {
    t.updateLastActiveTime()
    // 先发送数据长度
    length := uint32(len(data))
    if err := binary.Write(t.conn, binary.BigEndian, length); err != nil {
        return err
    }
    // 发送数据内容
    _, err := t.conn.Write(data)
    return err
}

// Receive 接收数据
func (t *TCPTransport) Receive() ([]byte, error) {
    t.updateLastActiveTime()
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

func (t *TCPTransport) Close() error {
    close(t.heartbeatStop)
    return t.conn.Close()
}