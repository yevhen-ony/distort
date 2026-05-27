package object

import (
	"context"
	t "dos/internal/common/types"
	"errors"
)

type CommandApplier interface {
  	Apply(context.Context, ObjectCommand) error
}

var (
	ErrUnknownObjectCommand = errors.New("unknown object command")
)


type LocalCommandApplier struct {
  	repo Repository 
}

func NewLocalCommandApplier(repo Repository) (*LocalCommandApplier, error) {
	if repo == nil {
		return nil, errors.New("missing object repository")
	}
	applier := &LocalCommandApplier{
		repo: repo,
	}
	return applier, nil
}

func (a *LocalCommandApplier) Apply(
  	ctx context.Context,
  	cmd ObjectCommand,
) error {

  	switch {
  	case cmd.CreateObject != nil:
  		c := cmd.CreateObject
  		return a.repo.Create(ctx, c.ObjectID, c.Replication)

	case cmd.DeleteObject != nil:
		c := cmd.DeleteObject
		return a.repo.Delete(ctx, c.ObjectID)

  	case cmd.SetReplication != nil:
  		c := cmd.SetReplication
  		return a.repo.SetReplication(ctx, c.ObjectID, c.Replication)

  	case cmd.AddChunk != nil:
  		c := cmd.AddChunk
		slot := t.ObjectSlot{
			ObjectID: c.ObjectID,
			ChunkKey: c.ChunkKey,
		}
  		return a.repo.AddChunk(ctx, slot, c.ChunkID)

  	case cmd.DeleteChunk != nil:
		c := cmd.DeleteChunk
  		return a.repo.DeleteChunk(ctx, t.ObjectSlot{
			ObjectID: c.ObjectID,
			ChunkKey: c.ChunkKey,
		})

  	default:
		return ErrUnknownObjectCommand
  	}
}
