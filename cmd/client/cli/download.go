package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
	"dos/cmd/client/render"
	"dos/internal/services/client/domain/progress"
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

			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer app.Close()

 			if cfg.OutputFormat() == "text" {
				cancel := app.Presenter.RunLoop(ctx)
				defer cancel()
  			}

			app.App.SetOnProgress(func(p *progress.ObjectProgress){
				app.Presenter.Update(p)
			})	
			if err = app.App.Download(ctx, objectID, destPath); err != nil {
				app.Presenter.Update(render.NewErrorResult("download", err))
				return err
			}
			return nil
		},
	}
	downloadCmd.Flags().String("dest", "", "dest file or dir the object to be stored")
	return downloadCmd
}
