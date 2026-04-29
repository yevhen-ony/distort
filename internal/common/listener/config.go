package listener

import "fmt"

type ListenerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c ListenerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
