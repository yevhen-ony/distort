package client

type Chunk struct {
	ID       string
	Checksum string
	Data     []byte
}

type Target struct {
	ID   string
	Addr string
}
