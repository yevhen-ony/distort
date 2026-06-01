package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
	"dos/cmd/client/render"
)

func MakeSystemCmd(cfg *app.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use: "system",
		Short: "show system-level cluster information",
	}
	listCmd.AddCommand(
		MakeLeaderCmd(cfg),
		MakePingCmd(cfg),
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

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err 
			}
			defer app.Close()

			res, err := app.App.DiscoverMaster(ctx)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("discover_master", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil
		},
	}
	return listObjectsCmd
}

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

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err 
			}
			defer app.Close()


			res, err := app.App.Ping(ctx, addr)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("ping", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil
		},
	}
	return pingCmd 
}
