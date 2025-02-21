package codec

import (
	"io"

	"google.golang.org/protobuf/proto"
)

type ProtobufCodec struct{}

func NewProtobufCodec() *ProtobufCodec {
	return &ProtobufCodec{}
}

func (c *ProtobufCodec) Encode(w io.Writer, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidProtobufMessage
	}
	
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	
	_, err = w.Write(data)
	return err
}

func (c *ProtobufCodec) Decode(r io.Reader, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidProtobufMessage
	}
	
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	
	return proto.Unmarshal(data, message)
}

func (c *ProtobufCodec) ContentType() string {
	return "application/x-protobuf"
}