package client

import (
	"context"

	t "dos/internal/common/types"
)

type Transport interface {
	SendChunk(context.Context, t.NodeAccess, *Chunk) error
	ReceiveChunk(context.Context, t.NodeAccess, string) (*Chunk, error)

}
