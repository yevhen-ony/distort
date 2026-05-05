package progress

import (
	t "dos/internal/common/types"
	"fmt"
	"strings"
)

type TransferEventKind int

const (
	TransferObjectKind TransferEventKind = iota
	TransferChunkKind
)

type ProgressEvent struct {
	Object *ObjectState
	Chunk  *ChunkState
}

type ObjectState struct {
	ObjectID t.ObjectID
}

func (os ObjectState) ToEvent() *ProgressEvent {
	return &ProgressEvent{Object: &os}
}

type ChunkState struct {
	ObjectID t.ObjectID
	Key      t.ChunkKey
	Meta     t.ChunkMeta
	Node     t.NodeID
}

func (cs ChunkState) ToEvent() *ProgressEvent {
	return &ProgressEvent{Chunk: &cs}
}

type ProgressView struct {
	objectState ObjectState
	chunksOrder []t.ChunkKey
	chunkStates map[t.ChunkKey]ChunkState
}

func NewProgressView() *ProgressView {
	return &ProgressView{
		chunkStates: make(map[t.ChunkKey]ChunkState),
	}
}

func (p *ProgressView) Update(event *ProgressEvent) {
	if p == nil || event == nil {
		return
	}
	if event.Object != nil {
		p.updateObject(event.Object)
		return
	}
	if event.Chunk != nil {
		p.updateChunk(event.Chunk)
		return
	}
}

func (p *ProgressView) updateObject(state *ObjectState) {
	p.objectState = *state
}

func (p *ProgressView) updateChunk(state *ChunkState) {
	if _, ok := p.chunkStates[state.Key]; !ok {
		p.chunksOrder = append(p.chunksOrder, state.Key)
	}
	p.chunkStates[state.Key] = *state
}

func (p *ProgressView) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "OBJECT: %s\n", p.objectState.ObjectID)

	fmt.Fprintf(&b,
		"%-10s %-20s %-10s %10s\n",
		"KEY", "ID", "SIZE", "CHECKSUM",
	)
	for _, key := range p.chunksOrder {
		state := p.chunkStates[key]

		mb := float64(state.Meta.Digest.Size) / (1024 * 1024)
		fmt.Fprintf(&b,
			"%-10s %-20s %8.1f MB %10s\n",
			state.Key, state.Meta.ID, mb, state.Meta.Digest.Checksum)
	}
	return b.String()
}
