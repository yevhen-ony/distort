package main

import (
	"context"
	"dos/cmd/client/app"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func MakePingCmd(cfg *app.Config) *cobra.Command {
	pingCmd := &cobra.Command{
		Use: "ping [addr]",
		Short: "ping resource",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			if len(args) == 0 || args[0] == "" {
				return errors.New("missing addr")
			}
			addr := args[0]

			render, err := NewRender(&cfg.CLI)
			if err != nil {
				return fmt.Errorf("init render: %w", err)
			}

			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()

			var rerr error
			res, err := app.Ping(ctx, addr)
			if err != nil {
				rerr = render.Error("ping", err)
			} else {
				rerr = render.Ping(res)
			}
			if rerr != nil {
				return fmt.Errorf("render: %w", rerr)
			}

			return nil
		},
	}
	return pingCmd 
}
