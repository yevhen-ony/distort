package app

import (
	"dos/internal/common/config"
	"time"
)


type Config struct {
	Client ClientConfig `yaml:"client"`
}

type ClientConfig struct {
	MasterAddr string `yaml:"master_addr"`
	ChunkSize config.Size `yaml:"chunk_size"`
	FrameSize config.Size `yaml:"frame_size"`
	TransferConcurrency int `yaml:"concurrency"`
	RenderRefreshInterval time.Duration `yaml:"render_refresh"`
}

func (c *Config) MasterAddr() string {
	return c.Client.MasterAddr
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
