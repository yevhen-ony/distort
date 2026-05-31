package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dos/cmd/client/app"
	"dos/internal/common/config"
)

const (
	outputKey = "output"
)


func main() {
	if err := run(); err != nil {
		fmt.Println(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := &app.Config{}
	if err := config.LoadConfig("config.yml", &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	root := &cobra.Command{
		Use: "dos",
	}
	BindFlags(cfg, root)

	root.AddCommand(MakeUploadCmd(cfg))
	root.AddCommand(MakeDownloadCmd(cfg))
	root.AddCommand(MakeListCmd(cfg))
	root.AddCommand(MakeDescribeCmd(cfg))
	root.AddCommand(MakeScaleObjectCmd(cfg))
	root.AddCommand(MakePingCmd(cfg))
	root.AddCommand(MakeSystemCmd(cfg))

	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute: %w")
	}
	return nil
}


func ApplyFlags(config *app.Config, cmd *cobra.Command) error {

  	if config == nil {
  		return fmt.Errorf("missing config")
  	}

  	out, err := cmd.Flags().GetString(outputKey)
  	if err != nil {
  		return fmt.Errorf("read --%s: %w", outputKey, err)
  	}

  	switch out {
  	case "text", "json":
  		config.CLI.OutputFormat = out
  	default:
  		return fmt.Errorf("invalid --%s %q (allowed: text,json)", outputKey, out)
  	}
  	return nil
}


func BindFlags(config *app.Config, cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(outputKey, "o", "text", "output format: text|json")

}



