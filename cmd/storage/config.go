package main

import (
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"
)

type Config struct {
	API             api.ServerConfig          `yaml:"api"`
	Store           store.ChunkStorageConfig  `yaml:"store"`
	Listen          listener.ListenerConfig   `yaml:"listen"`
	MasterTransport transport.MasterConfig    `yaml:"master_transport"`
	ChunkTransport  chunkrpc.Config           `yaml:"chunk_transport"`
	Service         core.StorageServiceConfig `yaml:"service"`
	Logger          logger.LogConfig          `yaml:"logger"`
}
