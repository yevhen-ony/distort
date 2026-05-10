package listener

import "fmt"

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
