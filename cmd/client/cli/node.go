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
		Use:     "node",
		Aliases: []string{"n"},
		Short:   "node-related operations",
	}
	nodeCmd.AddCommand(
		MakeListNodesCmd(cfg),
		MakeInspectNodeCmd(cfg),
		MakeTriggerReportCmd(cfg),
	)

	return nodeCmd
}

func MakeListNodesCmd(cfg *app.Config) *cobra.Command {
	listNodesCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list storage nodes",
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

func MakeInspectNodeCmd(cfg *app.Config) *cobra.Command {
	inspectNodeCmd := &cobra.Command{
		Use:   "inspect [addr]",
		Short: "inspect storage node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			addr := args[0]

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer app.Close()

			res, err := app.App.InspectNode(ctx, addr)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("inspect_node", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil
		},
	}
	return inspectNodeCmd
}


func MakeTriggerReportCmd(cfg *app.Config) *cobra.Command {
  	triggerReportCmd := &cobra.Command{
  		Use:   "report [addr]",
  		Short: "trigger storage node report",
  		Args:  cobra.ExactArgs(1),
  		RunE: func(cmd *cobra.Command, args []string) error {
  			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
  			defer stop()

  			if err := ApplyFlags(cfg, cmd); err != nil {
  				return fmt.Errorf("apply config flags: %w", err)
  			}

			addr := args[0]

  			all, err := cmd.Flags().GetBool("all")
  			if err != nil {
  				return fmt.Errorf("all flag: %w", err)
  			}

  			chunkIDs, err := cmd.Flags().GetStringArray("chunk")
  			if err != nil {
  				return fmt.Errorf("chunk flag: %w", err)
  			}

  			a, err := RunApp(ctx, cfg)
  			if err != nil {
  				return err
  			}
  			defer a.Close()

  			res, err := a.App.TriggerReport(ctx, app.TriggerReportQuery{
  				Addr:     addr,
  				All:      all,
  				ChunkIDs: chunkIDs,
  			})
  			if err != nil {
  				return a.Presenter.Update(render.NewErrorResult("trigger_report", err))
  			}
  			return a.Presenter.Update(res)
  		},
  	}

  	triggerReportCmd.Flags().Bool("all", false, "stage and report all local chunks")
  	triggerReportCmd.Flags().StringArray("chunk", nil, "chunk id to stage and report; repeated")

  	return triggerReportCmd
}

