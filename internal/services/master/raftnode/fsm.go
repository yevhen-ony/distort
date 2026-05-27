package raftnode

import (
	"context"
	"dos/internal/services/master/domain/object"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type ObjectFSM struct {
	codec object.CommandCodec
	applier object.CommandApplier
}

func NewObjectFSM(codec object.CommandCodec, applier object.CommandApplier) (*ObjectFSM, error) {
	if codec == nil {
		return nil, fmt.Errorf("missing codec")
	}
	if applier == nil {
		return nil, fmt.Errorf("missing command applier")
	}
	fsm := &ObjectFSM{
		codec: codec,
		applier: applier,
	}
	return fsm, nil
}

func (f *ObjectFSM) Apply(log *raft.Log) interface{} {
  	cmd, err := f.codec.Decode(log.Data)
  	if err != nil {
  		return err
  	}

  	return f.applier.Apply(context.Background(), cmd)
}

func (f *ObjectFSM) Snapshot() (raft.FSMSnapshot, error) {
  	return &NopSnapshot{}, nil
}

func (f *ObjectFSM) Restore(rc io.ReadCloser) error {
  	defer rc.Close()
  	return nil
}

type NopSnapshot struct{}

func (s *NopSnapshot) Persist(sink raft.SnapshotSink) error {
  	return sink.Close()
}

func (s *NopSnapshot) Release() {}
