package progress

import (
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"fmt"
	"sync"
)

type ProgressHandler func(*ObjectProgress)

func NOPProgressHanler(*ObjectProgress) {}

type ObjectStatus string

const (
	ObjectInProgress ObjectStatus = "In progress"
	ObjectDone       ObjectStatus = "Done"
	ObjectFailed     ObjectStatus = "Failed"
)

type ObjectProgress struct {
	ObjectID    t.ObjectID                       `json:"object_id"`
	ChunksOrder []t.ChunkKey                     `json:"-"`
	Chunks      map[t.ChunkKey]chunkrpc.Progress `json:"chunks"`
	Status      ObjectStatus                     `json:"status"`
	Reason      string                           `json:"reason,omitempty"`

	mu sync.Mutex
}

func NewObjectProgress(objectID t.ObjectID) *ObjectProgress {
	return &ObjectProgress{
		ObjectID: objectID,
		Chunks:   make(map[t.ChunkKey]chunkrpc.Progress),
		Status:   ObjectInProgress,
	}
}

func (op *ObjectProgress) UpdateChunk(key t.ChunkKey, chunk chunkrpc.Progress) {
	op.mu.Lock()
	defer op.mu.Unlock()

	if _, ok := op.Chunks[key]; !ok {
		op.ChunksOrder = append(op.ChunksOrder, key)
	}
	op.Chunks[key] = chunk
}

func (op *ObjectProgress) Fail(reason string) {
	op.Status = ObjectFailed
	op.Reason = reason
}

func (op *ObjectProgress) Done() {
	op.Status = ObjectDone
}

func (op *ObjectProgress) GetStatusStr() string {
	if op.Reason == "" {
		return string(op.Status)
	}
	return fmt.Sprintf("%s -> %s", op.Status, op.Reason)
}
