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

func MakeScaleObjectCmd(cfg *app.Config) *cobra.Command {
	scaleObjectCmd := &cobra.Command{
		Use: "scale [object-id] --replicas [N]",
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
			
			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ScaleObjects(ctx, objectID, replicaCount)
		},
	}
	scaleObjectCmd.Flags().Int("replicas", -1, "desired replication count")
	return scaleObjectCmd
}
