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
	ChunkSize           config.Size   `yaml:"chunk_size"`
	FrameSize           config.Size   `yaml:"frame_size"`
	TransferConcurrency int           `yaml:"concurrency"`
	RPCTimeout          time.Duration `yaml:"rpc_timeout"`
}

type CLIConfig struct {
	OutputFormat          string        `yaml:"output"`
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
	return c.CLI.RenderRefreshInterval
}

func (c *Config) OutputFormat() string {
	return c.CLI.OutputFormat
}

func (c *Config) RPCTimeout() time.Duration {
	return c.Client.RPCTimeout
}
