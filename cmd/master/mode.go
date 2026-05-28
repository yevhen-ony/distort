package main

import (
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/object"
)

type MasterMode interface {
	MasterState() m.MasterState
	ObjectAuthority() *object.Authority
}

func InitMasterMode(config *Config) (MasterMode, error) {
	if config.RaftEnabled() {
		return NewMasterRaftMode(config)
	}
	return NewMasterLocalMode(config)
}
