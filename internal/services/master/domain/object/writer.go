package object

import (
	"context"
	t "dos/internal/common/types"
	"errors"
)

type ObjectWriterImpl struct {
	submit CommandSubmitter
}

func NewObjectWriterImpl(submit CommandSubmitter) (*ObjectWriterImpl, error) {
	if submit == nil {
		return nil, errors.New("missing command applier")
	}
	ow := &ObjectWriterImpl{
		submit: submit,
	}
	return ow, nil
}

func (ow *ObjectWriterImpl) Create(ctx context.Context, id t.ObjectID, repl int) error {
	cmd := CreateObjectCommand{
		ObjectID:    id,
		Replication: repl,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *ObjectWriterImpl) Delete(ctx context.Context, id t.ObjectID) error {
	cmd := DeleteObjectCommand{
		ObjectID: id,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *ObjectWriterImpl) SetReplication(ctx context.Context, id t.ObjectID, repl int) error {
	cmd := SetReplicationCommand{
		ObjectID:    id,
		Replication: repl,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *ObjectWriterImpl) AddChunk(ctx context.Context, slot t.ObjectSlot, chunkID t.ChunkID) error {
	cmd := AddChunkCommand{
		ObjectID: slot.ObjectID,
		ChunkKey: slot.ChunkKey,
		ChunkID:  chunkID,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}

func (ow *ObjectWriterImpl) DeleteChunk(ctx context.Context, slot t.ObjectSlot) error {
	cmd := DeleteChunkCommand{
		ObjectID: slot.ObjectID,
		ChunkKey: slot.ChunkKey,
	}
	return ow.submit.Submit(ctx, cmd.ToCommand())
}
