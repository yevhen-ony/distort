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
	Meta      t.ChunkMeta
	NodeRef   t.NodeRef
	SentBytes int64

	status ChunkStatus
	reason string
}

func NewProgress(meta t.ChunkMeta, nodeRef t.NodeRef) Progress {
	p := Progress{
		Meta:    meta,
		NodeRef: nodeRef,
		status:  ChunkInProgress,
	}
	return p
}

func (p *Progress) Fail(reason string) {
	p.status = ChunkFailed
	p.reason = reason
}

func (p *Progress) Done() {
	p.status = ChunkDone 
}

func (p *Progress) GetStatusStr() string {
	if p.reason == "" {
		return string(p.status)
	}
	return fmt.Sprintf("%s -> %s", p.status, p.reason)
}
