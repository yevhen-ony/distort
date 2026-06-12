package object

import (
	"context"
	"errors"
	"testing"

	t "dos/internal/common/types"

	"github.com/stretchr/testify/require"
)

func TestCommandBackedObjectWriter_SubmitsCommand(tt *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		call func(*CommandBackedObjectWriter) error
		want ObjectCommand
	}{
		"create_object": {
			call: func(writer *CommandBackedObjectWriter) error {
				return writer.Create(ctx, "object-1", 2)
			},
			want: (&CreateObjectCommand{
				ObjectID:    "object-1",
				Replication: 2,
			}).ToCommand(),
		},
		"delete_object": {
			call: func(writer *CommandBackedObjectWriter) error {
				return writer.Delete(ctx, "object-1")
			},
			want: (&DeleteObjectCommand{
				ObjectID: "object-1",
			}).ToCommand(),
		},
		"set_replication": {
			call: func(writer *CommandBackedObjectWriter) error {
				return writer.SetReplication(ctx, "object-1", 3)
			},
			want: (&SetReplicationCommand{
				ObjectID:    "object-1",
				Replication: 3,
			}).ToCommand(),
		},
		"add_chunk": {
			call: func(writer *CommandBackedObjectWriter) error {
				return writer.AddChunk(ctx, t.ObjectSlot{
					ObjectID: "object-1",
					ChunkKey: "000001",
				}, "chunk-1")
			},
			want: (&AddChunkCommand{
				ObjectID: "object-1",
				ChunkKey: "000001",
				ChunkID:  "chunk-1",
			}).ToCommand(),
		},
		"delete_chunk": {
			call: func(writer *CommandBackedObjectWriter) error {
				return writer.DeleteChunk(ctx, t.ObjectSlot{
					ObjectID: "object-1",
					ChunkKey: "000001",
				})
			},
			want: (&DeleteChunkCommand{
				ObjectID: "object-1",
				ChunkKey: "000001",
			}).ToCommand(),
		},
	}

	for name, test := range tests {
		tt.Run(name, func(tt *testing.T) {
			submit := &testCommandSubmitter{}
			writer, err := NewCommandBackedObjectWriter(submit)
			require.NoError(tt, err)

			err = test.call(writer)

			require.NoError(tt, err)
			require.Equal(tt, test.want, submit.cmd)
		})
	}
}

func TestCommandBackedObjectWriter_OnSubmitError(tt *testing.T) {
	wantErr := errors.New("submit failed")
	submit := &testCommandSubmitter{err: wantErr}

	writer, err := NewCommandBackedObjectWriter(submit)
	require.NoError(tt, err)

	err = writer.Create(context.Background(), "object-1", 2)
	require.ErrorIs(tt, err, wantErr)
}

type testCommandSubmitter struct {
	cmd ObjectCommand
	err error
}

func (s *testCommandSubmitter) Submit(_ context.Context, cmd ObjectCommand) error {
	if s.err != nil {
		return s.err
	}
	s.cmd = cmd
	return nil
}
