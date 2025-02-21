package protocol

// MessageCodec 定义消息编解码接口
type MessageCodec interface {
    // Encode 将消息编码为字节流
    Encode(message *Message) ([]byte, error)

    // Decode 将字节流解码为消息
    Decode(data []byte) (*Message, error)
}