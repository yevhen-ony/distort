package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
)

var (
	ErrDigestMismatch = errors.New("digest mismatch")
)

type Checksum string

type Digest struct {
	Checksum Checksum `json:"checksum"`
	Size     int64    `json:"size"`
}

func (d *Digest) Match(other *Digest) error {
	if other == nil {
		return fmt.Errorf("compare with nil: %w", ErrDigestMismatch)
	}
	if d.Size != other.Size {
		return fmt.Errorf(
			"size: got %d, want %d: %w",
			d.Size, other.Size, ErrDigestMismatch,
		)
	}
	if d.Checksum != other.Checksum {
		return fmt.Errorf(
			"checksum: got %s, want %s: %w",
			d.Checksum, other.Checksum, ErrDigestMismatch,
		)
	}
	return nil
}

func (d *Digest) Clone() Digest {
	if d == nil {
		return Digest{}
	}
	return Digest{
		Checksum: d.Checksum,
		Size:     d.Size,
	}
}

type Digester struct {
	hash  hash.Hash
	total int64
}

func New() *Digester {
	return &Digester{
		hash: sha256.New(),
	}
}

func (d *Digester) Write(p []byte) (int, error) {
	n, err := d.hash.Write(p)
	if n > 0 {
		d.total += int64(n)
	}
	return n, err
}

func (d *Digester) Digest() Digest {
	return Digest{
		Size:     d.total,
		Checksum: d.Checksum(),
	}
}

func (d *Digester) Checksum() Checksum {
	sum := d.hash.Sum(nil)
	enc := hex.EncodeToString(sum)
	return Checksum(enc)
}
