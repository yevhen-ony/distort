package chunkrpc

type TransferFeatures struct {
	onProgress ProgressHandler
}

type SessionOption func(*TransferFeatures)

func WithProgress(h ProgressHandler) SessionOption {
	return func(s *TransferFeatures) {
		s.onProgress = h
	}
}

func applySessionOptions(opts []SessionOption) TransferFeatures {
	features := TransferFeatures{}
	for _, opt := range opts {
		opt(&features)
	}
	return features
}
