package main

import (
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
)

type Config struct {
	API     api.ServerConfig                `yaml:"api"`
	Store   store.ChunkStorageConfig        `yaml:"store"`
	Listen  ListenerConfig                  `yaml:"listen"`
	Master  transport.MasterTransportConfig `yaml:"master"`
	Service core.StorageServiceConfig       `yanl:"service"`
}

type ListenerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c ListenerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func loadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	err := k.Load(file.Provider(path), yaml.Parser())
	if err != nil {
		return nil, fmt.Errorf("load yaml: %w", err)
	}

	convert := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "__", ".")
		return s
	}

	err = k.Load(env.Provider("", ".", convert), nil)
	if err != nil {
		return nil, fmt.Errorf("load env: %w", err)
	}

	cfg := &Config{}
	conf := koanf.UnmarshalConf{Tag: "yaml"}
	err = k.UnmarshalWithConf("", cfg, conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return cfg, nil

}
