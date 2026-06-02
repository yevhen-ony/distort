package main

import (
	"context"
	"dos/cmd/client/app"
	"dos/cmd/client/render"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

const (
	replicasKey = "replicas"
)

func MakeObjectCmd(cfg *app.Config) *cobra.Command {
	objectCmd := &cobra.Command{
		Use:     "object",
		Short:   "object-related operations",
		Aliases: []string{"o"},
	}
	objectCmd.AddCommand(
		MakeDescribeObjectCmd(cfg),
		MakeScaleObjectCmd(cfg),
		MakeListObjectsCmd(cfg),
		MakeCreateObjectCmd(cfg),
	)

	return objectCmd
}

func MakeDescribeObjectCmd(cfg *app.Config) *cobra.Command {
	descObjectCmd := &cobra.Command{
		Use:     "object [object-id]",
		Aliases: []string{"o"},
		Short:   "describe object",
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

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer app.Close()

			res, err := app.App.DescribeObject(ctx, objectID)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("describe_object", err))
			} else {
				app.Presenter.Update(res)
			}
			return nil
		},
	}
	return descObjectCmd
}

func MakeScaleObjectCmd(cfg *app.Config) *cobra.Command {
	scaleObjectCmd := &cobra.Command{
		Use:   "scale [object-id]",
		Short: "scale replication object",
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
			replicaCount, err := cmd.Flags().GetInt("replicas")
			if err != nil {
				return fmt.Errorf("replicas flag: %w", err)
			}
			if replicaCount < 0 {
				return fmt.Errorf("missing or invalid replica count")
			}

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer app.Close()

			if err = app.App.ScaleObject(ctx, objectID, replicaCount); err != nil {
				app.Presenter.Update(render.NewErrorResult("scale_object", err))
			}
			return nil
		},
	}
	scaleObjectCmd.Flags().Int(replicasKey, -1, "desired replication count")
	scaleObjectCmd.MarkFlagRequired(replicasKey)
	return scaleObjectCmd
}

func MakeListObjectsCmd(cfg *app.Config) *cobra.Command {
	listObjectsCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list objects",
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

func MakeCreateObjectCmd(cfg *app.Config) *cobra.Command {
	createObjectCmd := &cobra.Command{
		Use:   "create [object-id]",
		Short: "create empty object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			objectID := args[0]

			a, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer a.Close()

			res, err := a.App.CreateObject(ctx, objectID)
			if err != nil {
				return a.Presenter.Update(render.NewErrorResult("create_object", err))
			}
			return a.Presenter.Update(res)
		},
	}
	return createObjectCmd
}
