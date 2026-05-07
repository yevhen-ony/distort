package domain

import (
	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"fmt"
	"strings"
	"sync"
)

type ObjectProgress struct {
	ObjectID t.ObjectID
	ChunksOrder []t.ChunkKey
	Chunks map[t.ChunkKey]chunkrpc.Progress 
	Done bool

	mu sync.Mutex
}

func NewObjectProgress(objectID t.ObjectID) *ObjectProgress {
	return &ObjectProgress{
		ObjectID: objectID,
		Chunks: make(map[t.ChunkKey]chunkrpc.Progress),
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


func (op *ObjectProgress) String() string {
  	op.mu.Lock()
  	defer op.mu.Unlock()

  	var b strings.Builder
  	fmt.Fprintf(&b, "OBJECT: %s\n", op.ObjectID)

  	fmt.Fprintf(
  		&b,
  		"%-10s %-20s %-10s %-10s %-6s\n",
  		"KEY", "ID", "SIZE", "SENT", "DONE",
  	)

  	for _, key := range op.ChunksOrder {
  		ch, ok := op.Chunks[key]
  		if !ok {
  			continue
  		}

  		var sizeMB float64
  		if ch.Meta.Digest != nil {
  			sizeMB = float64(ch.Meta.Digest.Size) / (1024 * 1024)
  		}
  		sentMB := float64(ch.SentBytes) / (1024 * 1024)

  		fmt.Fprintf(
  			&b,
  			"%-10s %-20s %8.1fMB %8.1fMB %-6t\n",
  			key, ch.Meta.ID, sizeMB, sentMB, ch.Done,
  		)
  	}

  	return b.String()
}

