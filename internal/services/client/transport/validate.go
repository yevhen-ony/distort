package transport

import (
	"fmt"

	t "dos/internal/common/types"
)

func ReceiveChunkValidate(target t.NodeRef, chunkID t.ChunkID) error {
	if err := validateNodeAccess(&target); err != nil {
		return err
	}
	if err := validateChunkID(chunkID); err != nil {
		return err
	}
	return nil
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
