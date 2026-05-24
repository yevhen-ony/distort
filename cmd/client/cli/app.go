package main

import (
	"errors"
	"fmt"

	"dos/internal/common/connect"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/services/client/transport"
)

type App struct {
	Config *Config

	Conn             *connect.ConnCache
	MasterTransport  *transport.MasterTransport
	StorageTransport *chunkrpc.Transport
}

func (app *App) Close() error {
	if app == nil || app.Conn == nil {
		return nil
	}
	return app.Conn.Close()
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	if config == nil {
		return nil, errors.New("missing config")
	}

	masterT, err := transport.NewMasterTransport(conn, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init master transport: %w", err)
	}

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init storage transport: %w", err)
	}

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init service: %w", err)
	}

	app := &App{
		Config:           config,
		Conn:             conn,
		MasterTransport:  masterT,
		StorageTransport: chunkT,
	}
	return app, nil
}
