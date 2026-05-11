package main

import (
	"dos/internal/common/config"
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"time"
)

type Config struct {
	Logger  logger.Config   `yaml:"logger"`
	Listen  listener.Config `yaml:"listen"`
	Service ServiceConfig   `yaml:"service"`
}

type ServiceConfig struct {
	ReplicationCount      int           `yaml:"replication_count"`
	ChunkAllocationMargin config.Size   `yaml:"chunk_allocation_margin"`
	NodeInactivityTimeout time.Duration `yaml:"node_inactivity_timeout"`
	NodeCleanupInterval   time.Duration `yaml:"node_cleanup_interval"`
	ReconcileQueueLength  int           `yanl:"reconcile_queue_length"`
}

func (c *Config) ReplicationCount() int {
	return c.Service.ReplicationCount
}

func (c *Config) ChunkAllocationMarginBytes() int64 {
	return int64(c.Service.ChunkAllocationMargin)
}

func (c *Config) NodeCleanupInterval() time.Duration {
	return c.Service.NodeCleanupInterval
}

func (c *Config) NodeInactivityTimeout() time.Duration {
	return c.Service.NodeInactivityTimeout
}

func  (c *Config) ReconcileQueueLength() int {
	return c.Service.ReconcileQueueLength
}
