package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CodecTestSuite struct {
	suite.Suite
	codec MessageCodec
}

func (s *CodecTestSuite) SetupTest() {
	s.codec = NewDefaultCodec()
}

func (s *CodecTestSuite) TestEncodeAndDecode() {
	// 准备测试数据
	header := &Header{
		ID:          1,
		Type:        TypeRequest,
		Compress:    0,
		Codec:       "json",
		ServiceName: "UserService",
		MethodName:  "GetUser",
		Metadata: map[string]string{
			"trace_id": "123456",
		},
		Timeout: time.Second * 2,
	}

	message := &Message{
		Header: header,
		Data:   []byte("test data"),
	}

	// 测试编码
	encoded, err := s.codec.Encode(message)
	s.NoError(err)
	s.NotEmpty(encoded)

	// 测试解码
	decoded, err := s.codec.Decode(encoded)
	s.NoError(err)
	s.NotNil(decoded)

	// 验证解码结果
	s.Equal(message.Header.ID, decoded.Header.ID)
	s.Equal(message.Header.Type, decoded.Header.Type)
	s.Equal(message.Header.Compress, decoded.Header.Compress)
	s.Equal(message.Header.Codec, decoded.Header.Codec)
	s.Equal(message.Header.ServiceName, decoded.Header.ServiceName)
	s.Equal(message.Header.MethodName, decoded.Header.MethodName)
	s.Equal(message.Header.Metadata, decoded.Header.Metadata)
	s.Equal(message.Header.Timeout, decoded.Header.Timeout)
	s.Equal(message.Data, decoded.Data)
}

func (s *CodecTestSuite) TestDecodeInvalidData() {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "空数据",
			data:    []byte{},
			wantErr: ErrInvalidMessage,
		},
		{
			name:    "数据长度不足",
			data:    make([]byte, 8),
			wantErr: ErrInvalidMessage,
		},
		{
			name:    "魔数错误",
			data:    make([]byte, 12),
			wantErr: ErrInvalidMagic,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.codec.Decode(tt.data)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *CodecTestSuite) TestEncodeEmptyMessage() {
	message := &Message{
		Header: &Header{},
		Data:   nil,
	}

	encoded, err := s.codec.Encode(message)
	s.NoError(err)
	s.NotEmpty(encoded)

	decoded, err := s.codec.Decode(encoded)
	s.NoError(err)
	s.NotNil(decoded)
	s.Empty(decoded.Data)
}

func TestCodecSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}