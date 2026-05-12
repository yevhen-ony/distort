package main

import (
	"dos/internal/common/config"

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
