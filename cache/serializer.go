package cache

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// JSONSerializer JSON 序列化器
type JSONSerializer struct{}

func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MsgpackSerializer Msgpack 序列化器（更高效）
type MsgpackSerializer struct{}

func (s *MsgpackSerializer) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (s *MsgpackSerializer) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}
