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

func MakeNodeCmd(cfg *app.Config) *cobra.Command {
	nodeCmd := &cobra.Command{
		Use: "node",
		Aliases: []string{"n"},
		Short: "node-related operations",
	}
	nodeCmd.AddCommand(
		MakeListNodesCmd(cfg),
	)

	return nodeCmd
}


func MakeListNodesCmd(cfg *app.Config) *cobra.Command {
	listNodesCmd := &cobra.Command{
		Use: "list",
		Aliases: []string{"ls"},
		Short: "list storage nodes",
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
