package object

import (
	"errors"

	m "dos/internal/services/master"
)

type Authority struct {
	m.ObjectReader
	m.ObjectWriter
}

type AuthorityDeps struct {
	Reader m.ObjectReader
	Writer m.ObjectWriter
}

func NewAuthority(deps AuthorityDeps) (*Authority, error) {
	if deps.Reader == nil {
		return nil, errors.New("missing object reader")
	}
	if deps.Writer == nil {
		return nil, errors.New("missing object writer")
	}
	oa := &Authority{
		ObjectReader: deps.Reader,
		ObjectWriter: deps.Writer,
	}
	return oa, nil
}
