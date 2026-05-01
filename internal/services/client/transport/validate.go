package transport

import (
	"fmt"

	t "dos/internal/common/types"
	c "dos/internal/services/client"
)


func SendChunkValidate(target t.NodeRef, chunk *c.Chunk) error {
	if err := validateNodeAccess(&target); err != nil {
		return err
	}
	if err := validateChunk(chunk); err != nil {
		return err
	}
	return nil
}

func ReceiveChunkValidate(target t.NodeRef, chunkID t.ChunkID) error {
	if err := validateNodeAccess(&target); err != nil {
		return err
	}
	if err := validateChunkID(chunkID); err != nil {
		return err
	}
	return nil
}

func validateChunk(chunk *c.Chunk) error {
	if chunk == nil {
		return fmt.Errorf("nil chunk: %w", ErrInputInvalid)
	}
	if len(chunk.Data) == 0 {
		return fmt.Errorf("empty chunk data: %w", ErrInputInvalid)
	}
	return validateChunkID(chunk.Meta.ID)
}

func validateChunkID(chunkID t.ChunkID) error {
	if chunkID == "" {
		return fmt.Errorf("empty chunk id: %w", ErrInputInvalid)
	}
	return nil
}

func validateNodeAccess(node *t.NodeRef) error {
	if node.ID == "" {
		return fmt.Errorf("empty target id: %w", ErrInputInvalid)
	}
	if node.Addr == "" {
		return fmt.Errorf("empty target addr: %w", ErrInputInvalid)
	}
	return nil
}
