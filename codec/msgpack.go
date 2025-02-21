package codec

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

type MsgpackCodec struct{}

func NewMsgpackCodec() *MsgpackCodec {
	return &MsgpackCodec{}
}

func (c *MsgpackCodec) Encode(w io.Writer, v interface{}) error {
	enc := msgpack.NewEncoder(w)
	return enc.Encode(v)
}

func (c *MsgpackCodec) Decode(r io.Reader, v interface{}) error {
	dec := msgpack.NewDecoder(r)
	return dec.Decode(v)
}

func (c *MsgpackCodec) ContentType() string {
	return "application/x-msgpack"
}