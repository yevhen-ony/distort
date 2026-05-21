package main

import (
	"time"

	"dos/internal/common/config"
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/common/metrics/prom"
)

type Config struct {
	Logger    logger.Config   `yaml:"logger"`
	Listen    listener.Config `yaml:"listen"`
	Metrics   prom.Config     `yaml:"metrics"`
	Store     StoreConfig     `yaml:"store"`
	Transport TransportConfig `yaml:"transport"`
	Service   ServiceConfig   `yaml:"service"`
}

type StoreConfig struct {
	MaxStorage config.Size `yaml:"max_storage"`
	RootDir    string      `yaml:"root_dir"`
}

type TransportConfig struct {
	AdvertiseAddr string      `yaml:"advertise_addr"`
	MasterAddr    string      `yaml:"master_addr"`
	FrameSize     config.Size `yaml:"frame_size"`
}

type ServiceConfig struct {
	HeartbeatInterval   time.Duration `yaml:"heartbeat_interval"`
	RegistrationTimeout time.Duration `yaml:"registration_timeout"`
	ReplicationTimeout  time.Duration `yaml:"replication_timeout"`
	ReportInterval      time.Duration `yaml:"report_interval"`
	ReportQueueCapacity int           `yaml:"report_queue_capacity"`
	MaxParallelHeavyOps int           `yaml:"max_parallel_heavy_ops"`
}

func (cfg *Config) MaxStorageBytes() int64 {
	return int64(cfg.Store.MaxStorage)
}

func (cfg *Config) StorageRootDir() string {
	return cfg.Store.RootDir
}

func (cfg *Config) AdvertiseAddr() string {
	return cfg.Transport.AdvertiseAddr
}

func (cfg *Config) MasterAddr() string {
	return cfg.Transport.MasterAddr
}

func (cfg *Config) FrameSize() int64 {
	return int64(cfg.Transport.FrameSize)
}

func (cfg *Config) HeartbeatInterval() time.Duration {
	return cfg.Service.HeartbeatInterval
}

func (cfg *Config) ReportInterval() time.Duration {
	return cfg.Service.ReportInterval
}

func (cfg *Config) RegistrationTimeout() time.Duration {
	return cfg.Service.RegistrationTimeout
}

func (cfg *Config) ReplicationTimeout() time.Duration {
	return cfg.Service.ReplicationTimeout
}

func (cfg *Config) QueueCapacity() int {
	return cfg.Service.ReportQueueCapacity
}

func (cfg *Config) MaxParallelHeavyOps() int {
	return cfg.Service.MaxParallelHeavyOps
}
