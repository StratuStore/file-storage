package queue

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type GobMarshaler struct{}

func (g *GobMarshaler) Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *GobMarshaler) Unmarshal(data []byte, v interface{}) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("gob decode failed: %w", err)
	}
	return nil
}
