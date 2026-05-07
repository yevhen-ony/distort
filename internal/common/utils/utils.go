package utils

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
