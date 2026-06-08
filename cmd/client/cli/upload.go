package main

import (
	"context"
	"dos/cmd/client/app"
	"dos/cmd/client/render"
	"dos/internal/services/client/domain/progress"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
)

func MakeUploadCmd(cfg *app.Config) *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "upload [path]",
		Aliases: []string{"ul"},
		Short: "upload file to the object storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			path := args[0]
			if err := EnsureFileExists(path); err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			objectID, err := cmd.Flags().GetString("id")
			if err != nil {
				return fmt.Errorf("read id flag: %w", err)
			}
			if objectID == "" {
				objectID = filepath.Base(path)
			}

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			a, err := RunApp(ctx, cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer a.Close()


 			if cfg.OutputFormat() == "text" {
				cancel := a.Presenter.RunLoop(ctx)
				defer cancel()
  			}

			a.App.SetOnProgress(func(p *progress.ObjectProgress) {
				a.Presenter.Update(p)
			})
			err = a.App.Upload(ctx, objectID, path)
			if err != nil {
				a.Presenter.Update(render.NewErrorResult("upload", err))
				return err
			}
			return nil
		},
	}
	pushCmd.Flags().String("id", "", "object id of the file being pushed")
	return pushCmd
}
