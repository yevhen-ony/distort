package transport

import "dos/internal/common/config"

type StorageTransportConfig struct {
	FrameSize config.Size `yaml:"frame_size"`
}

type MasterTransportConfig struct {
	Addr string `yaml:"addr"`
}
