package digest
import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	t "dos/internal/common/types"
)

type Digest struct {
	Checksum t.Checksum
	Size int64 
}

func (d Digest) Equal(o *Digest) bool {
	return d.Size == o.Size &&
		d.Checksum == o.Checksum
}

func (d *Digest) Clone() *Digest {
	return &Digest{
		Checksum: d.Checksum,
		Size: d.Size,
	}
}

type Digester struct {
	hash hash.Hash
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
		Size: d.total,
		Checksum: d.Checksum(),
	}
}

func (d *Digester) Checksum() t.Checksum {
	sum := d.hash.Sum(nil)
	enc := hex.EncodeToString(sum)
	return t.Checksum(enc)
}	
