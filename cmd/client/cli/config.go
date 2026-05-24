package main

import (
	"dos/internal/common/config"
	"time"

	"github.com/spf13/cobra"
)

var (
	masterAddrKey = "master-addr"
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

func (cfg *Config) BindFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(masterAddrKey, "", "master address")
}

func (cfg *Config) ApplyFlags(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}

	if flag := cmd.Flag(masterAddrKey); flag != nil && flag.Changed {
		v, err := cmd.Flags().GetString(masterAddrKey)
		if err != nil {
			return err
		}
		cfg.Client.MasterAddr = v
	}
	return nil
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
