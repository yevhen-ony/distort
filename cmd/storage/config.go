package main

import (
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/services/storage/api"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/store"
	"dos/internal/services/storage/transport"
	"dos/internal/services/storage/worker"
)

type Config struct {
	API       api.ServerConfig                `yaml:"api"`
	Store     store.ChunkStorageConfig        `yaml:"store"`
	Listen    listener.ListenerConfig         `yaml:"listen"`
	Master    transport.MasterTransportConfig `yaml:"master"`
	Service   core.StorageServiceConfig       `yaml:"service"`
	Logger    logger.LogConfig                `yaml:"logger"`
	Heartbeat worker.PeriodicConfig           `yaml:"heartbeat"`
}
