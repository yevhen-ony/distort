package raftnode

import (
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

type Config struct {
	NodeID    string `yaml:"node_id"`
	BindAddr  string `yaml:"bind_addr"`
	DataDir   string `yaml:"data_dir"`
	Bootstrap bool   `yaml:"bootstrap"`

	ApplyTimeout      time.Duration `yaml:"apply_timeout"`
	SnapshotInterval  time.Duration `yaml:"snapshot_interval"`
	SnapshotThreshold uint64        `yaml:"snapshot_threshold"`
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("missing config")
	}

	if c.NodeID == "" {
		return fmt.Errorf("missing raft node id")
	}
	if c.BindAddr == "" {
		return fmt.Errorf("missing raft bind addr")
	}
	return nil
}

func (c *Config) RaftConfig() *raft.Config {
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(c.NodeID)

    if c.SnapshotInterval > 0 {
        cfg.SnapshotInterval = c.SnapshotInterval
    }
    if c.SnapshotThreshold > 0 {
        cfg.SnapshotThreshold = c.SnapshotThreshold
    }
	
	return cfg
}
