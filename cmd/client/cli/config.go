package main

import (
	"dos/internal/services/client/io/file"
	"dos/internal/services/client/transport"

	"github.com/spf13/cobra"
)

var (
	masterAddrKey = "master-addr"
)

type Config struct {
	Master  transport.MasterTransportConfig  `yaml:"master"`
	Storage transport.StorageTransportConfig `yaml:"storage"`
	Chunker file.FileChunkerConfig           `yaml:"chunker"`
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
		cfg.Master.Addr = v
	}
	return nil
}
