package main

import (
	"dos/internal/common/config"
	"dos/internal/common/logger"
	mresolve "dos/internal/common/master/resolve"
	"dos/internal/common/metrics/prom"
	"dos/internal/services/master/raftnode"
	"fmt"
	"time"
)

type Config struct {
	Master  mresolve.Config `yaml:"master"`
	Logger  logger.Config   `yaml:"logger"`
	Metrics prom.Config     `yaml:"metrics"`
	Raft    raftnode.Config `yaml:"raft"`
	Service ServiceConfig   `yaml:"service"`
}

type ServiceConfig struct {
	FrameSize                  config.Size   `yaml:"frame_size"`
	ChunkAllocationMargin      config.Size   `yaml:"chunk_allocation_margin_bytes"`
	ReplicationCount           int           `yaml:"replication_count"`
	ReplicationQueueLength     int           `yaml:"replication_queue_length"`
	ReplicationExecInterval    time.Duration `yaml:"replication_interval"`
	ReplicationPlannerInterval time.Duration `yaml:"replication_planner_interval"`
	NodeInactivityTimeout      time.Duration `yaml:"node_inactivity_timeout"`
	NodeCleanupInterval        time.Duration `yaml:"node_cleanup_interval"`
	CatalogCleanupInterval     time.Duration `yaml:"catalog_cleanup_interval"`
	ChunkStaleAfter            time.Duration `yaml:"chunk_stale_after"`
	RPCTimeout                 time.Duration `yaml:"rpc_timeout"`
}

func (c *Config) FrameSize() int64 {
	return int64(c.Service.FrameSize)
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

func (c *Config) ReplicationQueueLength() int {
	return c.Service.ReplicationQueueLength
}

func (c *Config) CatalogCleanupInterval() time.Duration {
	return c.Service.CatalogCleanupInterval
}

func (c *Config) ReplicationExecInterval() time.Duration {
	return c.Service.ReplicationExecInterval
}

func (c *Config) ChunkStaleAfter() time.Duration {
	return c.Service.ChunkStaleAfter
}

func (c *Config) ReplicationPlannerInterval() time.Duration {
	return c.Service.ReplicationPlannerInterval
}

func (c *Config) RaftEnabled() bool {
	return c.Raft.Enable
}

func (c *Config) RPCTimeout() time.Duration {
	return c.Service.RPCTimeout
}

func (c *Config) ListeningAddr() string {
	return fmt.Sprintf("0.0.0.0:%d", c.Master.APIPort)
}

