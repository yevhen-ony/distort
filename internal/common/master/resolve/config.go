package resolve

import (
	"errors"
	"fmt"
	"slices"
)

type Config struct {
	APIPort  int      `yaml:"api_port"`
	RaftPort int      `yaml:"raft_port"`
	Self     string   `yaml:"self"`
	Peers    []string `yaml:"peers"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("missing config")
	}
	if c.APIPort == 0 {
		return errors.New("missing api port")
	}
	if len(c.Peers) == 0 {
		return errors.New("missing peers")
	}

	seen := make(map[string]struct{}, len(c.Peers))
	for _, peer := range c.Peers {
		if peer == "" {
			return errors.New("empty peer")
		}
		if _, ok := seen[peer]; ok {
			return fmt.Errorf("duplicate peer id %q", peer)
		}
		seen[peer] = struct{}{}
	}

	return nil
}

func (c *Config) ValidateWithRaft() error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Self == "" {
		return errors.New("missing self")
	}
	if !slices.Contains(c.Peers, c.Self) {
		return fmt.Errorf("self missing from peers: %q, peers = %v, len = %d", c.Self, c.Peers, len(c.Peers))
	}
	if c.RaftPort == 0 {
		return errors.New("missing raft port")
	}
	return nil
}
