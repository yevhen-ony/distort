package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"dos/internal/common/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := &Config{}
	if err := config.LoadConfig("config.yml", &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	root := &cobra.Command{
		Use: "dos",
	}
	cfg.BindFlags(root)

	root.AddCommand(MakePushCmd(cfg))
	root.AddCommand(MakePullCmd(cfg))
	root.AddCommand(MakeListCmd(cfg))


	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w")
	}
	return nil
}

func MakePushCmd(cfg *Config) *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "push [path]",
		Short: "push file to the object storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			path := args[0]
			objectID, err := cmd.Flags().GetString("object-id")
			if err != nil {
				return fmt.Errorf("read object-id flag: %w", err)
			}
			if objectID == "" {
				objectID = filepath.Base(path)
			}

			if err := cfg.ApplyFlags(cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()

			return app.Push(ctx, objectID, path)
		},
	}
	pushCmd.Flags().String("object-id", "", "object id of the file being pushed")
	return pushCmd
}

func MakePullCmd(cfg *Config) *cobra.Command {
	pullCmd := &cobra.Command{
		Use: "pull [object-id]",
		Short: "pull object from the object store",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			objectID := args[0]
			destPath, err := cmd.Flags().GetString("dest")
			if err != nil {
				return fmt.Errorf("read dest flag: %w", err)
			}
			
			if err := cfg.ApplyFlags(cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			
			return app.Pull(ctx, objectID, destPath)
		},
	}
	pullCmd.Flags().String("dest", "", "dest file or dir the object to be stored")
	return pullCmd
}

func MakeListCmd(cfg *Config) *cobra.Command {
	listCmd := &cobra.Command{
		Use: "list",
		Short: "list resources",
	}
	listCmd.AddCommand(
		MakeListObjectsCmd(cfg),
		MakeListChunksCmd(cfg),
	)

	return listCmd
}

func MakeListObjectsCmd(cfg *Config) *cobra.Command {
	listObjectsCmd := &cobra.Command{
		Use: "objects",
		Short: "list objects",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := cfg.ApplyFlags(cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ListObjects(ctx)
		},
	}
	return listObjectsCmd
}

func MakeListChunksCmd(cfg *Config) *cobra.Command {
	listChunksCmd := &cobra.Command{
		Use: "chunks",
		Short: "list chunks",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := cfg.ApplyFlags(cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			return app.ListChunks(ctx)
		},
	}
	return listChunksCmd
}

