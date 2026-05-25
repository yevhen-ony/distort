package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
)

func MakeDownloadCmd(cfg *app.Config) *cobra.Command {
	downloadCmd := &cobra.Command{
		Use: "download [object-id]",
		Aliases: []string{"dl"},
		Short: "download object from the object store",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			objectID := args[0]
			destPath, err := cmd.Flags().GetString("dest")
			if err != nil {
				return fmt.Errorf("dest flag: %w", err)
			}
			
			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := app.NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			
			_ = app.Download(ctx, objectID, destPath)
			return nil
		},
	}
	downloadCmd.Flags().String("dest", "", "dest file or dir the object to be stored")
	return downloadCmd
}
