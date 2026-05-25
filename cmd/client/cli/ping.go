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
	listObjectsCmd := &cobra.Command{
		Use: "ping [addr]",
		Short: "ping resource",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			addr := args[0]
			if addr == "" {
				return errors.New("missing addr")
			}

			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()

			app.Ping(ctx, addr)
			return nil
		},
	}
	return listObjectsCmd
}
