package main

import (
	"context"
	"errors"
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

	root.AddCommand(MakeUploadCmd(cfg))
	root.AddCommand(MakePullCmd(cfg))
	root.AddCommand(MakeListCmd(cfg))
	root.AddCommand(MakeScaleObjectCmd(cfg))

	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w")
	}
	return nil
}

func MakeUploadCmd(cfg *Config) *cobra.Command {
	pushCmd := &cobra.Command{
		Use:   "upload [path]",
		Aliases: []string{"ul"},
		Short: "upload file to the object storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			path := args[0]
			objectID, err := cmd.Flags().GetString("id")
			if err != nil {
				return fmt.Errorf("read id flag: %w", err)
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

			return app.Upload(ctx, objectID, path)
		},
	}
	pushCmd.Flags().String("id", "", "object id of the file being pushed")
	return pushCmd
}

func MakePullCmd(cfg *Config) *cobra.Command {
	pullCmd := &cobra.Command{
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
			
			if err := cfg.ApplyFlags(cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}

			app, err := NewApp(cfg)
			if err != nil {
				return fmt.Errorf("init app: %w", err)
			}
			defer app.Close()
			
			return app.Download(ctx, objectID, destPath)
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
		MakeListNodesCmd(cfg),
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

func MakeListNodesCmd(cfg *Config) *cobra.Command {
	listNodesCmd := &cobra.Command{
		Use: "nodes",
		Short: "list nodes",
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
			return app.ListNodes(ctx)
		},
	}
	return listNodesCmd
}

func MakeScaleObjectCmd(cfg *Config) *cobra.Command {
	scaleObjectCmd := &cobra.Command{
		Use: "scale [object-id] --replicas [N]",
		Short: "scale replication object",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := cfg.ApplyFlags(cmd); err != nil {
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
			
			app, err := NewApp(cfg)
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

