package repo

import ( 
	"context"

	m "dos/internal/services/master"
	t "dos/internal/common/types"

)

type InMemObjectRepo struct {
	objects map[t.ObjectID]*m.Object
}

func NewInMemObjectRepo() *InMemObjectRepo {
	return &InMemObjectRepo{
		objects: map[t.ObjectID]*m.Object{},
	}
}

func (o *InMemObjectRepo) Create(_ context.Context, oid t.ObjectID) error {
	if _, ok := o.objects[oid]; ok {
		return m.ErrObjectExists 
	}

	o.objects[oid] = &m.Object{
		ID: oid,
		Chunks: map[t.ChunkKey]t.ChunkID{},
	}
	return nil
}

func (o *InMemObjectRepo) Get(_ context.Context, oid t.ObjectID) (m.Object, error) {
	obj, ok := o.objects[oid]
	if !ok {
		return m.Object{}, m.ErrObjectNotFound
	}
	return *obj.Clone(), nil
}

func (o *InMemObjectRepo) List(_ context.Context) []t.ObjectItem {
	res := make([]t.ObjectItem, 0, len(o.objects))
	for _, obj := range o.objects {
		res = append(res, t.ObjectItem{
			ID: obj.ID,
			ChunkCount: len(obj.Chunks),
		})
	}
	return res
}

func (o *InMemObjectRepo) AddChunk(_ context.Context, oid t.ObjectID, key t.ChunkKey, cid t.ChunkID) error {
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
