package protocol

import "time"

// MessageType 消息类型
type MessageType byte

const (
	// 请求消息类型
	TypeRequest MessageType = iota
	// 响应消息类型
	TypeResponse
	// 心跳消息类型
	TypeHeartbeat
)

// Message RPC消息结构
type Message struct {
	// 消息头
	Header *Header
	// 消息体
	Data []byte
}

// Header 消息头
type Header struct {
	// 消息ID
	ID uint64
	// 消息类型
	Type MessageType
	// 压缩类型
	Compress uint8
	// 序列化类型
	Codec string
	// 服务名称
	ServiceName string
	// 方法名称
	MethodName string
	// 元数据
	Metadata map[string]string
	// 超时时间
	Timeout time.Duration
	// 错误信息
	Error string
}