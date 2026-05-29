package app

import (
	"time"

	mresolve "dos/internal/common/master/resolve"
	"dos/internal/common/config"
)


type Config struct {
	Client ClientConfig `yaml:"client"`
	Master mresolve.Config `yaml:"master"`
}

type ClientConfig struct {
	ChunkSize config.Size `yaml:"chunk_size"`
	FrameSize config.Size `yaml:"frame_size"`
	TransferConcurrency int `yaml:"concurrency"`
	RenderRefreshInterval time.Duration `yaml:"render_refresh"`
}


func (c *Config) ChunkSize() int64 {
	return int64(c.Client.ChunkSize)
}

func (c *Config) FrameSize() int64 {
	return int64(c.Client.FrameSize)
}

func (c *Config) TransferConcurrency() int {
	return c.Client.TransferConcurrency
}

func (c *Config) RenderRefreshInterval() time.Duration {
	return c.Client.RenderRefreshInterval
}
