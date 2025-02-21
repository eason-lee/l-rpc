package codec

import "io"

// Codec 定义序列化接口
type Codec interface {
	// Encode 将数据序列化写入 Writer
	Encode(w io.Writer, v interface{}) error
	
	// Decode 从 Reader 中读取并反序列化数据
	Decode(r io.Reader, v interface{}) error
	
	// ContentType 返回序列化方式的内容类型
	ContentType() string
}

// DefaultCodec 默认的序列化方式
var DefaultCodec = NewJSONCodec()

// GetCodec 根据内容类型获取对应的序列化器
func GetCodec(contentType string) Codec {
	switch contentType {
	case "application/json":
		return NewJSONCodec()
	case "application/x-protobuf":
		return NewProtobufCodec()
	case "application/x-msgpack":
		return NewMsgpackCodec()
	default:
		return DefaultCodec
	}
}