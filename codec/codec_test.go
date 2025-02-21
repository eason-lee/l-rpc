package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CodecTestSuite struct {
	suite.Suite
	testData *TestStruct
}

type TestStruct struct {
	Name string
	Age  int
	Tags []string
}

func (s *CodecTestSuite) SetupTest() {
	s.testData = &TestStruct{
		Name: "test",
		Age:  20,
		Tags: []string{"go", "rpc"},
	}
}

func (s *CodecTestSuite) TestJSONCodec() {
	codec := NewJSONCodec()
	
	// 测试编码
	buf := new(bytes.Buffer)
	err := codec.Encode(buf, s.testData)
	s.NoError(err)

	// 测试解码
	result := new(TestStruct)
	err = codec.Decode(buf, result)
	s.NoError(err)
	s.Equal(s.testData.Name, result.Name)
	s.Equal(s.testData.Age, result.Age)
	s.Equal(s.testData.Tags, result.Tags)
}

func (s *CodecTestSuite) TestMsgpackCodec() {
	codec := NewMsgpackCodec()
	
	// 测试编码
	buf := new(bytes.Buffer)
	err := codec.Encode(buf, s.testData)
	s.NoError(err)

	// 测试解码
	result := new(TestStruct)
	err = codec.Decode(buf, result)
	s.NoError(err)
	s.Equal(s.testData.Name, result.Name)
	s.Equal(s.testData.Age, result.Age)
	s.Equal(s.testData.Tags, result.Tags)
}

func (s *CodecTestSuite) TestGetCodec() {
	tests := []struct {
		name        string
		contentType string
		want        string
	}{
		{"JSON", "application/json", "application/json"},
		{"Protobuf", "application/x-protobuf", "application/x-protobuf"},
		{"Msgpack", "application/x-msgpack", "application/x-msgpack"},
		{"Default", "unknown", "application/json"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			codec := GetCodec(tt.contentType)
			s.Equal(tt.want, codec.ContentType())
		})
	}
}

func TestCodecSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}