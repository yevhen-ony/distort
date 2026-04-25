package client

import "context"

type Transport interface {
	SendChunk(context.Context, NodeAccess, *Chunk) error
	ReceiveChunk(context.Context, NodeAccess, string) (*Chunk, error)

}
