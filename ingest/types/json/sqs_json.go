// This package encapsulates the JSON library used by the sidecar query server.
//
// Instead of using the standard library's JSON library, we use jsoniter.
// jsoniter is a drop-in replacement for the standard library's JSON library.
// It is faster than the standard library's JSON library.
package json

import (
	"encoding/json"
	"io"

	jsoniter "github.com/json-iterator/go"
)

type (
	RawMessage = jsoniter.RawMessage

	Marshaler   = json.Marshaler
	Unmarshaler = json.Unmarshaler

	Decoder = jsoniter.Decoder
)

var (
	jsonLib = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Marshal(v interface{}) ([]byte, error) {
	return jsonLib.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsonLib.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) *Decoder {
	return jsonLib.NewDecoder(r)
}
