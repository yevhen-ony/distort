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

const (
	nodeIDKey   = "node-id"
	nodeAddrKey = "node-addr"
	destKey     = "dest"
	chunkKeyKey = "key"
	chunkIDKey  = "id"
)

func MakeChunkCmd(cfg *app.Config) *cobra.Command {
	chunkCmd := &cobra.Command{
		Use:   "chunk",
		Aliases: []string{"c"},
		Short: "chunk-related operations",
	}
	chunkCmd.AddCommand(
		MakeGetChunkCmd(cfg),
		MakeAllocChunkCmd(cfg),
		MakePushChunkCmd(cfg),
		MakeDescribeChunkCmd(cfg),
		MakeListChunksCmd(cfg),
	)

	return chunkCmd
}

func MakeGetChunkCmd(cfg *app.Config) *cobra.Command {
	getChunkCmd := &cobra.Command{
		Use:   "get [chunk-id]",
		Short: "download an individual chunk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			chunkID := args[0]

			nodeID, err := cmd.Flags().GetString(nodeIDKey)
			if err != nil {
				return fmt.Errorf("node-id: %w", err)
			}
			nodeAddr, err := cmd.Flags().GetString(nodeAddrKey)
			if err != nil {
				return fmt.Errorf("node-addr: %w", err)
			}
			dest, err := cmd.Flags().GetString(destKey)
			if err != nil {
				return fmt.Errorf("dest: %w", err)
			}

			a, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer a.Close()

			res, err := a.App.DownloadChunk(ctx, app.DownloadChunkQuery{
				ChunkID:  chunkID,
				NodeID:   nodeID,
				NodeAddr: nodeAddr,
				DestPath: dest,
			})
			if err != nil {
				a.Presenter.Update(render.NewErrorResult("get_chunk", err))
				return err
			}
			return a.Presenter.Update(res)
		},
	}
	getChunkCmd.Flags().String(destKey, "", "dest file or dir chunk to be stored")

	getChunkCmd.Flags().String(nodeIDKey, "", "id of storage node; required")
	getChunkCmd.MarkFlagRequired(nodeIDKey)

	getChunkCmd.Flags().String(nodeAddrKey, "", "address of storage node; required")
	getChunkCmd.MarkFlagRequired(nodeAddrKey)

	return getChunkCmd
}

func MakeAllocChunkCmd(cfg *app.Config) *cobra.Command {
	allocChunkCmd := &cobra.Command{
		Use:     "allocate [object-id]",
		Aliases: []string{"alloc"},
		Short:   "allocate chunk for the specified object",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			objectID := args[0]

			key, err := cmd.Flags().GetString(chunkKeyKey)
			if err != nil {
				return fmt.Errorf("key: %w", err)
			}

			a, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer a.Close()

			res, err := a.App.AllocateChunk(ctx, app.AllocateChunkQuery{
				ObjectID: objectID,
				ChunkKey: key,
			})
			if err != nil {
				a.Presenter.Update(render.NewErrorResult("allocate_chunk", err))
				return err 
			}
			return a.Presenter.Update(res)
		},
	}

	allocChunkCmd.Flags().String(chunkKeyKey, "", "key of the chunk; required")
	allocChunkCmd.MarkFlagRequired(chunkKeyKey)

	return allocChunkCmd
}

func MakePushChunkCmd(cfg *app.Config) *cobra.Command {
	pushChunkCmd := &cobra.Command{
		Use:   "push [filepath]",
		Short: "push chunk to storage node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			path := args[0]

			chunkID, err := cmd.Flags().GetString(chunkIDKey)
			if err != nil {
				return fmt.Errorf("%s: %w", chunkIDKey, err)
			}
			nodeID, err := cmd.Flags().GetString(nodeIDKey)
			if err != nil {
				return fmt.Errorf("%s: %w", nodeIDKey, err)
			}
			nodeAddr, err := cmd.Flags().GetString(nodeAddrKey)
			if err != nil {
				return fmt.Errorf("%s: %w", nodeAddrKey, err)
			}

			a, err := RunApp(ctx, cfg)
			if err != nil {
				return err
			}
			defer a.Close()

			res, err := a.App.PushChunk(ctx, app.PushChunkQuery{
				NodeID:   nodeID,
				NodeAddr: nodeAddr,
				ChunkID:  chunkID,
				Path:     path,
			})
			if err != nil {
				a.Presenter.Update(render.NewErrorResult("push_chunk", err))
				return err
			}
			return a.Presenter.Update(res)
		},
	}

	pushChunkCmd.Flags().String(nodeIDKey, "", "storage node id; required")
	pushChunkCmd.MarkFlagRequired(nodeIDKey)

	pushChunkCmd.Flags().String(nodeAddrKey, "", "storage node address; required")
	pushChunkCmd.MarkFlagRequired(nodeAddrKey)

	pushChunkCmd.Flags().String(chunkIDKey, "", "chunk id; required")
	pushChunkCmd.MarkFlagRequired(chunkIDKey)

	return pushChunkCmd
}

func MakeDescribeChunkCmd(cfg *app.Config) *cobra.Command {
	descChunkCmd := &cobra.Command{
		Use: "describe [chunk-id]",
		Aliases: []string{"desc"},
		Short: "describe chunk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := ApplyFlags(cfg, cmd); err != nil {
				return fmt.Errorf("apply config flags: %w", err)
			}
			
			chunkID := args[0]
			
			app, err := RunApp(ctx, cfg)
			if err != nil {
				return err 
			}
			defer app.Close()

			res, err := app.App.DescribeChunk(ctx, chunkID)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("describe_chunk", err))
				return err
			}
			return app.Presenter.Update(res)
		},
	}
	return descChunkCmd
}

func MakeListChunksCmd(cfg *app.Config) *cobra.Command {
	listChunksCmd := &cobra.Command{
		Use: "list",
		Aliases: []string{"ls"},
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
			
			res, err := app.App.ListChunks(ctx)
			if err != nil {
				app.Presenter.Update(render.NewErrorResult("list_chunks", err))
				return err
			}
			return app.Presenter.Update(res)
		},
	}
	return listChunksCmd
}
