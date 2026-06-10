package resolve

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func validConfig() *Config {
	return &Config{
		APIPort:  10000,
		RaftPort: 12000,
		Self:     "master-1",
		Peers:    []string{"master-1", "master-2"},
	}
}

func TestConfig_Validate(tt *testing.T) {
	tt.Run("valid", func(tt *testing.T) {
		require.NoError(tt, validConfig().Validate())
	})

	tt.Run("nil", func(tt *testing.T) {
		var cfg *Config
		require.ErrorContains(tt, cfg.Validate(), "missing config")
	})

	tt.Run("missing api port", func(tt *testing.T) {
		cfg := validConfig()
		cfg.APIPort = 0

		require.ErrorContains(tt, cfg.Validate(), "missing api port")
	})

	tt.Run("missing peers", func(tt *testing.T) {
		cfg := validConfig()
		cfg.Peers = nil

		require.ErrorContains(tt, cfg.Validate(), "missing peers")
	})

	tt.Run("empty peer", func(tt *testing.T) {
		cfg := validConfig()
		cfg.Peers = []string{"master-1", ""}

		require.ErrorContains(tt, cfg.Validate(), "empty peer")
	})

	tt.Run("duplicate peer", func(tt *testing.T) {
		cfg := validConfig()
		cfg.Peers = []string{"master-1", "master-1"}

		require.ErrorContains(tt, cfg.Validate(), "duplicate peer id")
	})
}

func TestConfig_ValidateWithRaft(tt *testing.T) {
	tt.Run("valid", func(tt *testing.T) {
		require.NoError(tt, validConfig().ValidateWithRaft())
	})

	tt.Run("missing self", func(tt *testing.T) {
		cfg := validConfig()
		cfg.Self = ""

		require.ErrorContains(tt, cfg.ValidateWithRaft(), "missing self")
	})

	tt.Run("self missing from peers", func(tt *testing.T) {
		cfg := validConfig()
		cfg.Self = "master-3"

		require.ErrorContains(tt, cfg.ValidateWithRaft(), "self missing from peers")
	})

	tt.Run("missing raft port", func(tt *testing.T) {
		cfg := validConfig()
		cfg.RaftPort = 0

		require.ErrorContains(tt, cfg.ValidateWithRaft(), "missing raft port")
	})
}
