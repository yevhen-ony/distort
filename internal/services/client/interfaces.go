package client

import "context"

type Transport interface {
	SendChunk(context.Context, Target, *Chunk) error
	ReceiveChunk(context.Context, Target, string) (*Chunk, error)
}
