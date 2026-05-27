package resolve

import (
	t "dos/internal/common/types"
	"errors"
	"fmt"
)

type Config struct {
	Self  t.MasterID `yaml:"self"`
	Peers []Peer     `yaml:"peers"`
}

type Peer struct {
	ID       t.MasterID `yaml:"id"`
	APIAddr  string     `yaml:"api_addr"`
	RaftAddr string     `yaml:"raft_addr"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("missing config")
	}
	if len(c.Peers) == 0 {
		return errors.New("missing peers")
	}

	seen := make(map[t.MasterID]struct{}, len(c.Peers))
	for _, peer := range c.Peers {
		if peer.ID == "" {
			return errors.New("missing peer id")
		}
		if peer.APIAddr == "" {
			return fmt.Errorf("missing api addr for peer %q", peer.ID)
		}
		if _, ok := seen[peer.ID]; ok {
			return fmt.Errorf("duplicate peer id %q", peer.ID)
		}
		seen[peer.ID] = struct{}{}
	}

	if c.Self != "" {
		if _, ok := seen[c.Self]; !ok {
			return fmt.Errorf("missing peer for self %q", c.Self)
		}
	}

	return nil
}

func (c *Config) ValidateWithRaft() error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Self == "" {
		return ErrMissingSelf
	}

	for _, peer := range c.Peers {
		if peer.RaftAddr == "" {
			return fmt.Errorf("missing raft addr for peer %q", peer.ID)
		}
	}

	return nil
}
