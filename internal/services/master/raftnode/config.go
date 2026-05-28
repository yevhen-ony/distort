package raftnode

import (
	t "dos/internal/common/types"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

type Config struct {
	Enable bool `yaml:"enable"`

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
	return nil
}

func (c *Config) RaftConfig(id t.MasterID) *raft.Config {
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(id)

	if c.SnapshotInterval > 0 {
		cfg.SnapshotInterval = c.SnapshotInterval
	}
	if c.SnapshotThreshold > 0 {
		cfg.SnapshotThreshold = c.SnapshotThreshold
	}

	return cfg
}
