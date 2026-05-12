package repo

import (
	"context"
	"sync"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

type InMemObjectRepo struct {
	objects map[t.ObjectID]*m.Object
	mu sync.RWMutex
}

func NewInMemObjectRepo() *InMemObjectRepo {
	return &InMemObjectRepo{
		objects: map[t.ObjectID]*m.Object{},
	}
}

func (o *InMemObjectRepo) Create(_ context.Context, oid t.ObjectID, desiredReplication int) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, ok := o.objects[oid]; ok {
		return m.ErrObjectExists 
	}

	o.objects[oid] = &m.Object{
		ID: oid,
		Chunks: map[t.ChunkKey]t.ChunkID{},
		DesiredReplication: desiredReplication,
	}
	return nil
}

func (o *InMemObjectRepo) Get(_ context.Context, oid t.ObjectID) (m.Object, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	obj, ok := o.objects[oid]
	if !ok {
		return m.Object{}, m.ErrObjectNotFound
	}
	return *obj.Clone(), nil
}

func (o *InMemObjectRepo) GetReplication(_ context.Context, oid t.ObjectID) (int, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	obj, ok := o.objects[oid]
	if !ok {
		return 0, m.ErrObjectNotFound
	}
	return obj.DesiredReplication, nil
}

func (o *InMemObjectRepo) List(_ context.Context) []m.Object {
	o.mu.RLock()
	defer o.mu.RUnlock()

	res := make([]m.Object, 0, len(o.objects))
	for _, obj := range o.objects {
		res = append(res, *obj.Clone())
	}
	return res
}

func (o *InMemObjectRepo) AddChunk(_ context.Context, oid t.ObjectID, key t.ChunkKey, cid t.ChunkID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

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
