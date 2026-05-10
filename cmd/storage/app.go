package main

import (
	"dos/internal/common/connect"
	"dos/internal/services/storage/core"
	"dos/internal/services/storage/transport"
	"fmt"
)

type App struct {
	master *transport.Master
	identity *core.IdentityService
	report *core.ReportService
	storage *core.StorageService
}

func NewApp(cfg *Config) (*App, error) {
	conn := connect.NewConnCache()

	master, err := transport.NewMaster(conn, cfg)
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	identity := core.NewIdentityService(master, cfg)

	reportQueue := core.NewReportQueue()
}
