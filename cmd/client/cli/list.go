package main

import (
	"context"
	"dos/cmd/client/app"
	"dos/cmd/client/render"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func MakeListCmd(cfg *app.Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use: "list",
		Aliases: []string{"ls"},
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
		Aliases: []string{"o"},
		Short: "list objects",
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

			res, err := app.App.ListObjects(ctx)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("list_objects", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil 
		},
	}
	return listObjectsCmd
}

func MakeListChunksCmd(cfg *app.Config) *cobra.Command {
	listChunksCmd := &cobra.Command{
		Use: "chunks",
		Aliases: []string{"c"},
		Short: "list chunks",
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
			
			var rerr error
			res, err := app.App.ListChunks(ctx)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("list_chunks", err))
			} else {
				app.Presenter.Update(res)
			}
			return rerr
		},
	}
	return listChunksCmd
}

func MakeListNodesCmd(cfg *app.Config) *cobra.Command {
	listNodesCmd := &cobra.Command{
		Use: "nodes",
		Aliases: []string{"n"},
		Short: "list nodes",
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

			res, err := app.App.ListNodes(ctx)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("list_nodes", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil
		},
	}
	return listNodesCmd
}
