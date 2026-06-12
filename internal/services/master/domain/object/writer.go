package object

import (
	"context"
	t "dos/internal/common/types"
	"errors"
)

type CommandBackedObjectWriter struct {
	submit CommandSubmitter
}

func NewCommandBackedObjectWriter(submit CommandSubmitter) (*CommandBackedObjectWriter, error) {
	if submit == nil {
		return nil, errors.New("missing command applier")
	}
	ow := &CommandBackedObjectWriter{
		submit: submit,
	}
	return ow, nil
}

func (ow *CommandBackedObjectWriter) Create(ctx context.Context, id t.ObjectID, repl int) error {
	cmd := CreateObjectCommand{
		ObjectID:    id,
		Replication: repl,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *CommandBackedObjectWriter) Delete(ctx context.Context, id t.ObjectID) error {
	cmd := DeleteObjectCommand{
		ObjectID: id,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *CommandBackedObjectWriter) SetReplication(ctx context.Context, id t.ObjectID, repl int) error {
	cmd := SetReplicationCommand{
		ObjectID:    id,
		Replication: repl,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *CommandBackedObjectWriter) AddChunk(ctx context.Context, slot t.ObjectSlot, chunkID t.ChunkID) error {
	cmd := AddChunkCommand{
		ObjectID: slot.ObjectID,
		ChunkKey: slot.ChunkKey,
		ChunkID:  chunkID,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *CommandBackedObjectWriter) DeleteChunk(ctx context.Context, slot t.ObjectSlot) error {
	cmd := DeleteChunkCommand{
		ObjectID: slot.ObjectID,
		ChunkKey: slot.ChunkKey,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}
