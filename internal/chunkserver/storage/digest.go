package storage

import (
	"crypto/sha256"
	cs "dos/internal/chunkserver"
	"encoding/hex"
	"hash"
)


type Digester struct {
	h hash.Hash
	size int64
}

func NewDigester() *Digester {
	return &Digester{
		h: sha256.New(),
	}
}

func (d *Digester) Write(p []byte) (int, error) {
	n, err := d.h.Write(p)
	if n > 0 {
		d.size += int64(n)
	}
	return n, err
}

func (mt *Digester) Digest() cs.ChunkDigest {
	return cs.ChunkDigest{
		Size: mt.size,
		Checksum: hex.EncodeToString(mt.h.Sum(nil)),
	}
}


