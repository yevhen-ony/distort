package main

import (
	"context"
	"dos/cmd/client/app"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func MakeListCmd(cfg *app.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use: "list",
		Short: "list resources",
	}
	listCmd.AddCommand(
		MakeListObjectsCmd(cfg),
		MakeListChunksCmd(cfg),
		MakeListNodesCmd(cfg),
	)

	return listCmd
}

func MakeListObjectsCmd(cfg *app.Config) *cobra.Command {
	listObjectsCmd := &cobra.Command{
		Use: "objects",
		Short: "list objects",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ListObjects(ctx)
		},
	}
	return listObjectsCmd
}

func MakeListChunksCmd(cfg *app.Config) *cobra.Command {
	listChunksCmd := &cobra.Command{
		Use: "chunks",
		Short: "list chunks",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ListChunks(ctx)
		},
	}
	return listChunksCmd
}

func MakeListNodesCmd(cfg *app.Config) *cobra.Command {
	listNodesCmd := &cobra.Command{
		Use: "nodes",
		Short: "list nodes",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ListNodes(ctx)
		},
	}
	return listNodesCmd
}
