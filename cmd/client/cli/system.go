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
	systemCmd := &cobra.Command{
		Use:   "system",
		Short: "show system-level cluster information",
	}
	systemCmd.AddCommand(
		MakeLeaderCmd(cfg),
		MakePingCmd(cfg),
	)

	return systemCmd
}

func MakeLeaderCmd(cfg *app.Config) *cobra.Command {
	leaderCmd := &cobra.Command{
		Use:   "leader",
		Short: "master node leader information",
	}
	leaderCmd.AddCommand(
		MakeShowLeaderCmd(cfg),
		MakeTransferLeaderCmd(cfg),
	)

	return leaderCmd
}

func MakeShowLeaderCmd(cfg *app.Config) *cobra.Command {
	listObjectsCmd := &cobra.Command{
		Use:   "show",
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
				return err
			}
			app.Presenter.Update(res)
			return nil
		},
	}
	return listObjectsCmd
}

func MakeTransferLeaderCmd(cfg *app.Config) *cobra.Command {
	transferLeaderCmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer master leadership",
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

			if err = app.App.TransferLeadership(ctx); err != nil {
				app.Presenter.Update(render.NewErrorResult("transfer_leadership", err))
				return err
			}
			return nil
		},
	}
	return transferLeaderCmd
}

func MakePingCmd(cfg *app.Config) *cobra.Command {
	pingCmd := &cobra.Command{
		Use:   "ping [addr]",
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
				return err
			}
			app.Presenter.Update(res)
			return nil
		},
	}
	return pingCmd
}
