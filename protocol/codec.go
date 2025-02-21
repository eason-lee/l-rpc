package protocol

import (
	"encoding/binary"
)

// MessageCodec 消息编解码接口
type MessageCodec interface {
	// Encode 将消息编码为字节流
	Encode(message *Message) ([]byte, error)
	
	// Decode 将字节流解码为消息
	Decode(data []byte) (*Message, error)
}

// DefaultCodec 默认编解码器
type DefaultCodec struct{}

func NewDefaultCodec() *DefaultCodec {
	return &DefaultCodec{}
}

// 消息格式: 
// | 魔数 4字节 | 头部长度 4字节 | 消息长度 4字节 | 头部数据 | 消息数据 |
const (
	magicNumber = 0x11223344
)

func (c *DefaultCodec) Encode(message *Message) ([]byte, error) {
	// 编码消息头
	headerData, err := encode(message.Header)
	if err != nil {
		return nil, err
	}

	// 计算总长度
	totalLen := 12 + len(headerData) + len(message.Data)
	buf := make([]byte, totalLen)

	// 写入魔数
	binary.BigEndian.PutUint32(buf[0:4], magicNumber)
	// 写入头部长度
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(headerData)))
	// 写入消息体长度
	binary.BigEndian.PutUint32(buf[8:12], uint32(len(message.Data)))
	// 写入头部数据
	copy(buf[12:12+len(headerData)], headerData)
	// 写入消息数据
	copy(buf[12+len(headerData):], message.Data)

	return buf, nil
}

func (c *DefaultCodec) Decode(data []byte) (*Message, error) {
	if len(data) < 12 {
		return nil, ErrInvalidMessage
	}

	// 验证魔数
	magic := binary.BigEndian.Uint32(data[0:4])
	if magic != magicNumber {
		return nil, ErrInvalidMagic
	}

	// 获取头部和消息长度
	headerLen := binary.BigEndian.Uint32(data[4:8])
	dataLen := binary.BigEndian.Uint32(data[8:12])

	if uint32(len(data)) < 12+headerLen+dataLen {
		return nil, ErrInvalidMessage
	}

	// 解码消息头
	header := &Header{}
	err := decode(data[12:12+headerLen], header)
	if err != nil {
		return nil, err
	}

	// 构造消息
	message := &Message{
		Header: header,
		Data:   data[12+headerLen : 12+headerLen+dataLen],
	}

	return message, nil
}