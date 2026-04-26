package client

import (
	"context"

	t "dos/internal/common/types"
)

type Transport interface {
	SendChunk(context.Context, t.NodeRef, *Chunk) error
	ReceiveChunk(context.Context, t.NodeRef, string) (*Chunk, error)

}
