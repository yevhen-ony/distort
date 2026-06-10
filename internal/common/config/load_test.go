package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Service testServiceConfig `yaml:"service"`
}

type testServiceConfig struct {
	Name     string        `yaml:"name"`
	Enabled  bool          `yaml:"enabled"`
	Count    int           `yaml:"count"`
	Timeout  time.Duration `yaml:"timeout"`
	Peers    []string      `yaml:"peers"`
	MaxBytes Size          `yaml:"max_bytes"`
}

func defaultTestConfig() testConfig {
	return testConfig{
		Service: testServiceConfig{
			Name:     "storage",
			Enabled:  true,
			Count:    3,
			Timeout:  2 * time.Second,
			Peers:    []string{"master-1", "master-2"},
			MaxBytes: 10_000_000,
		},
	}
}

func assertConfigEqual(tt *testing.T, want testConfig, got testConfig) {
	tt.Helper()

	assert.Equal(tt, want.Service.Name, got.Service.Name, "service.name")
	assert.Equal(tt, want.Service.Enabled, got.Service.Enabled, "service.enabled")
	assert.Equal(tt, want.Service.Count, got.Service.Count, "service.count")
	assert.Equal(tt, want.Service.Timeout, got.Service.Timeout, "service.timeout")
	assert.Equal(tt, want.Service.Peers, got.Service.Peers, "service.peers")
	assert.Equal(tt, want.Service.MaxBytes, got.Service.MaxBytes, "service.max_bytes")
}

const yamlStr = `
service:
  name: storage
  enabled: true
  count: 3
  timeout: 2s
  peers:
    - master-1
    - master-2
  max_bytes: 10MB
`

func TestLoadConfig_LoadsYAML(tt *testing.T) {
	path := writeConfigFile(tt, yamlStr)

	var cfg testConfig

	err := LoadConfig(path, &cfg)
	require.NoError(tt, err, "load config failed")

	assertConfigEqual(tt, defaultTestConfig(), cfg)
}

func TestLoadConfig_EnvOverrides(tt *testing.T) {
	path := writeConfigFile(tt, yamlStr)

	tt.Run("array", func(tt *testing.T) {
		tt.Setenv("SERVICE__PEERS", "env-master-1,env-master-2")

		var cfg testConfig
		err := LoadConfig(path, &cfg)
		require.NoError(tt, err, "load config failed")

		expectedPeers := []string{"env-master-1", "env-master-2"}
		require.Equal(tt, expectedPeers, cfg.Service.Peers)
	})

	tt.Run("size", func(tt *testing.T) {
		tt.Setenv("SERVICE__MAX_BYTES", "3KB")

		var cfg testConfig
		err := LoadConfig(path, &cfg)
		require.NoError(tt, err, "load config failed")

		require.Equal(tt, int64(3_000), int64(cfg.Service.MaxBytes))
	})

	tt.Run("time", func(tt *testing.T) {
		tt.Setenv("SERVICE__TIMEOUT", "1m")

		var cfg testConfig
		err := LoadConfig(path, &cfg)
		require.NoError(tt, err, "load config failed")

		require.Equal(tt, time.Minute, cfg.Service.Timeout)
	})
}

func writeConfigFile(tt *testing.T, contents string) string {
	tt.Helper()

	path := tt.TempDir() + "/config.yml"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		tt.Fatalf("write config: %v", err)
	}
	return path
}
