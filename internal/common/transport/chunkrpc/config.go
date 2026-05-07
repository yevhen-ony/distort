package chunkrpc

import "dos/internal/common/config"

type Config struct {
	FrameSize config.Size `yaml:"frame_size"`
}
