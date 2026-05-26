package object

import (
	"context"
	"errors"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type ObjectReader interface {
	List(ctx context.Context) []m.Object
	Get(ctx context.Context, objectID t.ObjectID) (m.Object, error)
	Exists(ctx context.Context, objectID t.ObjectID) (bool, error)

	GetReplication(ctx context.Context, objectID t.ObjectID) (int, error)
	ExistsChunk(ctx context.Context, slot t.ObjectSlot) (bool, error)
	GetChunk(ctx context.Context, slot t.ObjectSlot) (t.ChunkID, error)
}

type ObjectWriter interface {
	Create(context.Context, t.ObjectID, int) error
	Delete(context.Context, t.ObjectID) error

	SetReplication(context.Context, t.ObjectID, int) error

	AddChunk(context.Context, t.ObjectSlot, t.ChunkID) error
	DeleteChunk(context.Context, t.ObjectSlot) error
}

type ObjectAuthority interface {
	ObjectReader
	ObjectWriter
}

type Authority struct {
	ObjectReader
	ObjectWriter
}

func NewObjectAuthority(reader ObjectReader, writer ObjectWriter) (*Authority, error) {
	if reader == nil {
		return nil, errors.New("missing object reader")
	}
	if writer == nil {
		return nil, errors.New("missing object writer")
	}
	oa := &Authority{
		ObjectReader: reader,
		ObjectWriter: writer,
	}
	return oa, nil
}

