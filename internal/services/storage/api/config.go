package api

import "dos/internal/common/config"

type ServerConfig struct {
	FrameSize config.Size 	`yaml:"frame_size"`
}
