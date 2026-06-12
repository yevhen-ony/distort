package object

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type CommandCodec interface {
	Encode(ObjectCommand) ([]byte, error)
	Decode([]byte) (ObjectCommand, error)
}

type JSONCommandCodec struct{}

func NewJSONCommandCodec() *JSONCommandCodec {
	return &JSONCommandCodec{}
}

func (s *JSONCommandCodec) Encode(cmd ObjectCommand) ([]byte, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(cmd)
}

func (s *JSONCommandCodec) Decode(data []byte) (ObjectCommand, error) {
	var cmd ObjectCommand

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&cmd); err != nil {
		return ObjectCommand{}, fmt.Errorf("decode object command: %w", err)
	}
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		return ObjectCommand{}, ErrInvalidObjectCommand
	}
	if err := cmd.Validate(); err != nil {
		return ObjectCommand{}, err
	}

	return cmd, nil
}
