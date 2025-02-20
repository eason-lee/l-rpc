package protocol

// Codec 定义编解码接口
type Codec interface {
    Encode(message *Message) ([]byte, error)
    Decode(data []byte) (*Message, error)
}