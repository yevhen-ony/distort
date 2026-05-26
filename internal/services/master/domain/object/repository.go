package object

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type Repository interface {
  	Create(context.Context, t.ObjectID, int) error
  	Delete(context.Context, t.ObjectID) error

  	List(context.Context) []m.Object
  	Get(context.Context, t.ObjectID) (m.Object, error)
  	Exists(context.Context, t.ObjectID) (bool, error)

  	GetReplication(context.Context, t.ObjectID) (int, error)
  	SetReplication(context.Context, t.ObjectID, int) error

  	ExistsChunk(context.Context, t.ObjectSlot) (bool, error)
  	AddChunk(context.Context, t.ObjectSlot, t.ChunkID) error
  	GetChunk(context.Context, t.ObjectSlot) (t.ChunkID, error)
  	DeleteChunk(context.Context, t.ObjectSlot) error
  }
