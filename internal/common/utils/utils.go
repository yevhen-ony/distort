package utils

import (
	"math/rand/v2"
	"time"
)

func SplitFrames(data []byte, size int64) [][]byte {
	if size <= 0 {
		panic("frame size must be positive")
	}
	l := int64(len(data)) 

	total := (l + size - 1) / size
	frames := make([][]byte, 0, total)

	for len(data) > 0 {
		n := min(size, int64(len(data)))
		frames = append(frames, data[:n])
		data = data[n:]
	}
	return frames
}

func Map[X any, Y any](xs []X, fn func(X)Y) []Y {
	ys := make([]Y, len(xs))
	for i, x := range xs {
		ys[i] = fn(x) 
	}
	return ys
}

func Select[X any](xs []X, keep func(X) bool) []X {
  	out := make([]X, 0, len(xs))
  	for _, x := range xs {
  		if keep(x) {
  			out = append(out, x)
  		}
  	}
  	return out
}

func RandomSelect[T any](ts []T, n int) []T {
	n = min(len(ts), n)
	perm := rand.Perm(len(ts))

	selected := make([]T, n)
	for i := range n {
		selected[i] = ts[perm[i]]
	}
	return selected
}

func RandomSelectOne[T any](ts []T) T {
	return ts[rand.IntN(len(ts))]
}

func Jitter(base time.Duration, frac float64) time.Duration {
	delta := float64(base) * frac
	j := (rand.Float64()*2 - 1) * delta
	return time.Duration(float64(base) + j)
}

