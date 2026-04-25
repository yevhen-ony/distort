package transport

type StorageTransportConfig struct {
	FrameSize int `yaml:"frame_size"`
}

type MasterTransportConfig struct {
	Addr string `yaml:"addr"`
}
