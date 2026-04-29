package main

import (
	"dos/internal/common/listener"
	"dos/internal/common/logger"
	"dos/internal/services/master/domain"
)

type Config struct {
	Logger  logger.LogConfig           `yaml:"logger"`
	Listen  listener.ListenerConfig    `yaml:"listen"`
	Service domain.MasterServiceConfig `yaml:"service"`
}
