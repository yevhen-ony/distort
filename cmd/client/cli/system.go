package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
)

func MakeSystemCmd(cfg *app.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use: "system",
		Short: "show system-level cluster information",
	}
	listCmd.AddCommand(
		MakeLeaderCmd(cfg),
	)

	return listCmd
}
func MakeLeaderCmd(cfg *app.Config) *cobra.Command {
	listObjectsCmd := &cobra.Command{
		Use: "leader",
		Short: "show current master leader",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := InitApp(cfg)
			if err != nil {
				return err 
			}
			defer app.Close()

			var rerr error 
			res, err := app.App.DiscoverMaster(ctx)
			if err != nil {
				rerr = app.Render.Error("leader", err)
			} else {
				rerr = app.Render.DiscoverMaster(res)
			}
			return rerr
		},
	}
	return listObjectsCmd
}
