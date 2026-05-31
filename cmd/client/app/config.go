package app

import (
	"time"

	"dos/internal/common/config"
	mresolve "dos/internal/common/master/resolve"
)

type Config struct {
	Client ClientConfig    `yaml:"client"`
	Master mresolve.Config `yaml:"master"`
	CLI    CLIConfig       `yaml:"cli"`
}

type ClientConfig struct {
	ChunkSize             config.Size   `yaml:"chunk_size"`
	FrameSize             config.Size   `yaml:"frame_size"`
	TransferConcurrency   int           `yaml:"concurrency"`
	RenderRefreshInterval time.Duration `yaml:"render_refresh"`
}

type CLIConfig struct {
	OutputFormat string `yaml:"output"`
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
