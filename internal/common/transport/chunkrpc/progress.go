package chunkrpc 

import t "dos/internal/common/types"

type ProgressHandler func(Progress)

type Progress struct {
	Meta        t.ChunkMeta
	NodeRef     t.NodeRef
	SentBytes   int64
	Done        bool
}

func NewProgress(meta t.ChunkMeta, nodeRef t.NodeRef) Progress {
	return Progress{
		Meta:    meta,
		NodeRef: nodeRef,
	}
}
