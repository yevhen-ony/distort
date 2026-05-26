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

func (o *InMemObjectRepo) Create(_ context.Context, oid t.ObjectID, replicas int) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, ok := o.objects[oid]; ok {
		return m.ErrObjectExists 
	}

	o.objects[oid] = &m.Object{
		ID: oid,
		Chunks: map[t.ChunkKey]t.ChunkID{},
		Replication: replicas,
	}
	return nil
}

func (o *InMemObjectRepo) Exists(_ context.Context, oid t.ObjectID) (bool, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	_, ok := o.objects[oid]
	return ok, nil
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
	return obj.Replication, nil
}

func (o *InMemObjectRepo) SetReplication(_ context.Context, oid t.ObjectID, count int) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if count < 0 {
		return m.ErrInvalidArgument
	}

	obj, ok := o.objects[oid]
	if !ok {
		return m.ErrObjectNotFound
	}
	obj.Replication = count
	return nil
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

func (o *InMemObjectRepo) ExistsChunk(_ context.Context, slot t.ObjectSlot) (bool, error) {
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	obj, ok := o.objects[slot.ObjectID]
	if !ok {
		return false, m.ErrObjectNotFound
	}
	_, ok = obj.Chunks[slot.ChunkKey]
	return ok, nil
}

func (o *InMemObjectRepo) AddChunk(_ context.Context, slot t.ObjectSlot, cid t.ChunkID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	obj, ok := o.objects[slot.ObjectID]
	if !ok {
		return m.ErrObjectNotFound
	}
	if _, ok := obj.Chunks[slot.ChunkKey]; ok {
		return m.ErrChunkKeyExists
	}
	obj.Chunks[slot.ChunkKey] = cid 
	return nil
}

func (o *InMemObjectRepo) GetChunk(_ context.Context, slot t.ObjectSlot) (t.ChunkID, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	obj, ok := o.objects[slot.ObjectID]
	if !ok {
		return "", m.ErrObjectNotFound
	}
	chunkID, ok := obj.Chunks[slot.ChunkKey]
	if !ok {
		return "", m.ErrChunkKeyNotFound
	}
	return chunkID, nil
}

func (o *InMemObjectRepo) DeleteChunk(_ context.Context, slot t.ObjectSlot) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	object, ok := o.objects[slot.ObjectID]
	if !ok {
		return m.ErrObjectNotFound 
	}
	_, ok = object.Chunks[slot.ChunkKey]
	if !ok {
		return m.ErrChunkKeyNotFound
	}

	delete(object.Chunks, slot.ChunkKey)
	return nil
}

func (o *InMemObjectRepo) Delete(_ context.Context, objectID t.ObjectID) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	obj, ok := o.objects[objectID]
	if !ok {
		return nil
	}
	
	if len(obj.Chunks) > 0 {
		return m.ErrObjectNotEmpty 
	}

	delete(o.objects, objectID)
	return nil

}
