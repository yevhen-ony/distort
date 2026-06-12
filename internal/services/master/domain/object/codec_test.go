package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONCommandCodec_RoundTrip(tt *testing.T) {
	codec := NewJSONCommandCodec()

	cmd := AddChunkCommand{
		ObjectID: "object-1",
		ChunkKey: "000001",
		ChunkID:  "chunk-1",
	}

	data, err := codec.Encode(cmd.ToCommand())
	require.NoError(tt, err)

	got, err := codec.Decode(data)
	require.NoError(tt, err)
	require.Equal(tt, cmd.ToCommand(), got)
}

func TestJSONCommandCodec_InvalidCommand(tt *testing.T) {
	codec := NewJSONCommandCodec()

	_, err := codec.Encode(ObjectCommand{})
	require.ErrorIs(tt, err, ErrInvalidObjectCommand)

	_, err = codec.Decode([]byte(`{}`))
	require.ErrorIs(tt, err, ErrInvalidObjectCommand)
}
