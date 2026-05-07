package types

import (
	"errors"
	"fmt"
)

func (m ChunkMeta) Match(other ChunkMeta) error {
	if other.ID != m.ID {
		return fmt.Errorf("id mismatch: %w", ErrChunkMetaMismatch)
	}

	if err := m.Digest.Match(other.Digest); err != nil {
		return errors.Join(err, ErrChunkMetaMismatch) 
	}

	return nil
}
