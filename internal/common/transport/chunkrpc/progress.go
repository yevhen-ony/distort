package chunkrpc 

import t "dos/internal/common/types"

type ProgressHandler func(Progress)

type Progress struct {
	Meta        t.ChunkMeta
	NodeRef     t.NodeRef
	SentBytes   int64
	Status string
}

func NewProgress(meta t.ChunkMeta, nodeRef t.NodeRef) Progress {
	p := Progress{
		Meta:    meta,
		NodeRef: nodeRef,
	}
	p.InProgress()
	return p
}

func (p *Progress) Fail() {
	p.Status = "Failed"
}

func (p *Progress) Done() {
	p.Status = "Done"
}

func (p *Progress) InProgress() {
	p.Status = "In progress"
}
