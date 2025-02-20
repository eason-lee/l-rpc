package protocol

// MessageType 定义消息类型
type MessageType byte

const (
    Request  MessageType = iota // 请求消息
    Response                    // 响应消息
)

// Message 定义 RPC 消息格式
type Message struct {
    Header Header
    Data   []byte
}

// Header 定义消息头
type Header struct {
    MessageType MessageType // 消息类型：请求/响应
    RequestID   uint64     // 请求ID，用于匹配请求和响应
    ServiceName string     // 服务名称
    MethodName  string     // 方法名称
    Error       string     // 错误信息，响应时使用
    DataLen     uint32     // 数据长度
}