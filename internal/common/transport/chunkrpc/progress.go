package chunkrpc

import (
	t "dos/internal/common/types"
	"fmt"
)

type ProgressHandler func(Progress)

type ChunkStatus string

const (
	ChunkInProgress ChunkStatus = "In progress"
	ChunkDone       ChunkStatus = "Done"
	ChunkFailed     ChunkStatus = "Failed"
)

type Progress struct {
	Meta      t.ChunkMeta `json:"chunk_meta"`
	NodeRef   t.NodeRef   `json:"node_ref"`
	SentBytes int64       `json:"sent_bytes"`

	Status ChunkStatus `json:"status"`
	Reason string      `json:"reason,omitempty"`
}

func NewProgress(meta t.ChunkMeta, nodeRef t.NodeRef) Progress {
	p := Progress{
		Meta:    meta,
		NodeRef: nodeRef,
		Status:  ChunkInProgress,
	}
	return p
}

func (p *Progress) Fail(reason string) {
	p.Status = ChunkFailed
	p.Reason = reason
}

func (p *Progress) Done() {
	p.Status = ChunkDone
}

func (p *Progress) GetStatusStr() string {
	if p.Reason == "" {
		return string(p.Status)
	}
	return fmt.Sprintf("%s -> %s", p.Status, p.Reason)
}
