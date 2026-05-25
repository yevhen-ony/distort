package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
)


func MakeDescribeCmd(cfg *app.Config) *cobra.Command {
	describeCmd := &cobra.Command{
		Use: "describe",
		Short: "describe resources",
	}
	describeCmd.AddCommand(
		MakeDescribeChunkCmd(cfg),
		MakeDescribeObjectCmd(cfg),
	)

	return describeCmd 
}

func MakeDescribeChunkCmd(cfg *app.Config) *cobra.Command {
	descChunkCmd := &cobra.Command{
		Use: "chunk [chunk-id]",
		Aliases: []string{"c"},
		Short: "describe chunk",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			chunkID := args[0]
			if chunkID == "" {
				return errors.New("missing chunk id")
			}
			
			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.DescribeChunk(ctx, chunkID)
		},
	}
	return descChunkCmd
}

func MakeDescribeObjectCmd(cfg *app.Config) *cobra.Command {
	descObjectCmd := &cobra.Command{
		Use: "object [object-id]",
		Aliases: []string{"o"},
		Short: "describe object",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			objectID := args[0]
			if objectID == "" {
				return errors.New("missing object id")
			}
			
			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.DescribeObject(ctx, objectID)
		},
	}
	return descObjectCmd
}

