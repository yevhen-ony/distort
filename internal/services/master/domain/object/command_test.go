package object

import (
	t "dos/internal/common/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestObjectCommand_Validate(tt *testing.T) {

	tt.Run("valid_command", func(tt *testing.T) {
		deleteCmd := (&DeleteObjectCommand{ObjectID: "object-1"}).ToCommand()
		require.NoError(tt, deleteCmd.Validate())
	})
	

	tt.Run("empty_command_invalid", func(tt *testing.T) {
		emptyCmd := ObjectCommand{}
		require.ErrorIs(tt, emptyCmd.Validate(), ErrInvalidObjectCommand)
	})

	tt.Run("mixed_command_invalid", func(tt *testing.T) {
		delCmd := &DeleteObjectCommand{ObjectID: "object-1"}
		addCmd := &AddChunkCommand{
			ObjectID: t.ObjectID("object-1"),
			ChunkKey: "000001",
			ChunkID: "chunk-1",
		}
		cmd := ObjectCommand{
			DeleteObject: delCmd,
			AddChunk: addCmd,
		}
		require.ErrorIs(tt, cmd.Validate(), ErrInvalidObjectCommand)
	})
}
