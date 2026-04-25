package transport

type NodeTransportConfig struct {
	FrameSize int `yaml:"frame_size"`
}

type MasterTransportConfig struct {
	Addr string `yaml:"addr"`
}
