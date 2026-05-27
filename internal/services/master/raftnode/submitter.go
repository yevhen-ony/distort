package raftnode

import (
	"context"
	"fmt"
	"time"

	"dos/internal/services/master/domain/object"

	"github.com/hashicorp/raft"
)

type CommandSubmitterDeps struct {
	Codec   object.CommandCodec
	Raft    *raft.Raft
	Timeout time.Duration
}

type CommandSubmitter struct {
	codec   object.CommandCodec
	raft    *raft.Raft
	timeout time.Duration
}

func NewCommandSubmitter(deps CommandSubmitterDeps) (*CommandSubmitter, error) {
	if deps.Codec == nil {
		return nil, fmt.Errorf("missing codec")
	}
	if deps.Raft == nil {
		return nil, fmt.Errorf("missing raft")
	}

	submitter := &CommandSubmitter{
		codec:   deps.Codec,
		raft:    deps.Raft,
		timeout: deps.Timeout,
	}
	return submitter, nil
}

func (s *CommandSubmitter) Submit(_ context.Context, cmd object.ObjectCommand) error {
	data, err := s.codec.Encode(cmd)
	if err != nil {
		return fmt.Errorf("command encode: %w", err)
	}

	f := s.raft.Apply(data, s.timeout)
	if err := f.Error(); err != nil {
		return err
	}
	if err, ok := f.Response().(error); ok {
		return err
	}

	return nil
}
