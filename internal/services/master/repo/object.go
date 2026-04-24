package repo

import ( 
	"context"

	m "dos/internal/services/master"

)

type InMemObjectRepo struct {
	objects map[m.ObjectID]*m.Object
}

func NewInMemObjectRepo() *InMemObjectRepo {
	return &InMemObjectRepo{
		objects: map[m.ObjectID]*m.Object{},
	}
}

func (o *InMemObjectRepo) Create(_ context.Context, oid m.ObjectID) error {
	if _, ok := o.objects[oid]; ok {
		return m.ErrObjectExists 
	}

	o.objects[oid] = &m.Object{
		ID: oid,
		Chunks: map[m.ChunkKey]m.ChunkID{},
	}
	return nil
}

func (o *InMemObjectRepo) Get(_ context.Context, oid m.ObjectID) (m.Object, error) {
	obj, ok := o.objects[oid]
	if !ok {
		return m.Object{}, m.ErrObjectNotFound
	}
	return *obj.Clone(), nil
}

func (o *InMemObjectRepo) AddChunk(_ context.Context, oid m.ObjectID, key m.ChunkKey, cid m.ChunkID) error {
	obj, ok := o.objects[oid]
	if !ok {
		return m.ErrObjectNotFound
	}
	if _, ok := obj.Chunks[key]; ok {
		return m.ErrChunkKeyExists
	}
	obj.Chunks[key] = cid 
	return nil
}
