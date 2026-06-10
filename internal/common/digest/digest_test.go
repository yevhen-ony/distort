package digest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDigester_Size(tt *testing.T) {
	dg := New()

	n, err := dg.Write([]byte("hello"))
	require.NoError(tt, err)
	require.Equal(tt, 5, n)

	got := dg.Digest()

	require.Equal(tt, int64(5), got.Size)
	require.NotEmpty(tt, got.Checksum)
}

func TestDigester_MultipleWrites(tt *testing.T) {
	oneShot := New()
	_, err := oneShot.Write([]byte("hello world"))
	require.NoError(tt, err)

	chunked := New()
	_, err = chunked.Write([]byte("hello "))
	require.NoError(tt, err)
	_, err = chunked.Write([]byte("world"))
	require.NoError(tt, err)

	require.Equal(tt, oneShot.Digest(), chunked.Digest())
}

func TestDigester_DifferentContent(tt *testing.T) {
	left := New()
	_, err := left.Write([]byte("hello"))
	require.NoError(tt, err)

	right := New()
	_, err = right.Write([]byte("world"))
	require.NoError(tt, err)

	require.NotEqual(tt, left.Digest().Checksum, right.Digest().Checksum)
	require.Equal(tt, left.Digest().Size, right.Digest().Size)
}

func TestDigest_Match(tt *testing.T) {
	d := Digest{Checksum: "abc", Size: 3}

	require.NoError(tt, d.Match(&Digest{Checksum: "abc", Size: 3}), "match expected")

	err := d.Match(nil)
	require.ErrorIs(tt, err, ErrDigestMismatch, "mismatch expected")

	err = d.Match(&Digest{Checksum: "abc", Size: 4})
	require.ErrorIs(tt, err, ErrDigestMismatch, "mismatch expected")

	err = d.Match(&Digest{Checksum: "def", Size: 3})
	require.ErrorIs(tt, err, ErrDigestMismatch, "mismatch expected")
}
