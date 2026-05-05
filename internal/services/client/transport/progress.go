package transport

import t "dos/internal/common/types"

type ChunkProgressHandler func(ChunkProgress)

type ChunkProgress struct {
	Meta        t.ChunkMeta
	NodeRef     t.NodeRef
	SentBytes   int64
	Done        bool
}

func NewChunkProgress(meta t.ChunkMeta, nodeRef t.NodeRef) ChunkProgress {
	return ChunkProgress{
		Meta:    meta,
		NodeRef: nodeRef,
	}
}
