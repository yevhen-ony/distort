package object

import (
	"context"
	"errors"
)

type CommandSubmitter interface {
	Submit(context.Context, ObjectCommand) error
}

type LocalCommandSubmitter struct {
	apply CommandApplier 
}

func NewLocalCommandSubmitter(apply CommandApplier) (*LocalCommandSubmitter, error) {
	if apply == nil {
		return nil, errors.New("missing command applier") 
	}

	submit := &LocalCommandSubmitter{
		apply: apply,
	}
	return submit, nil
}

func (s *LocalCommandSubmitter) Submit(ctx context.Context, cmd ObjectCommand) error {
	return s.apply.Apply(ctx, cmd)
}
