package message

import (
    "bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
)

// Codec defines the interface used to encode and decode messages.
type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	String() string
}


// protoCodec is a Codec implemetation with protobuf.
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) String() string {
	return "proto"
}
