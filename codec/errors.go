package codec

import "errors"

var (
	ErrInvalidProtobufMessage = errors.New("value is not a protobuf message")
	ErrUnsupportedCodec      = errors.New("unsupported codec type")
)