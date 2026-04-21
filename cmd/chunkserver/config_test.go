package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("yaml", func (t *testing.T) {
		cfg, err := loadConfig("./config.yml")
		require.NoError(t, err, "load config")

		assert.Equal(t, 5, cfg.API.PartSize)
		assert.Equal(t, "./chunkserver_data", cfg.Store.RootDir)
	})

	t.Run("env", func(t *testing.T) {
		err := os.Setenv("LISTEN__PORT", "80")
		require.NoError(t, err, "set env")	

		cfg, err := loadConfig("./config.yml")	
		require.NoError(t, err, "load config")

		assert.Equal(t, 80, cfg.Listen.Port)
		assert.Equal(t, "*", cfg.Listen.Host)
		assert.Equal(t, "./chunkserver_data", cfg.Store.RootDir)
	})
}
