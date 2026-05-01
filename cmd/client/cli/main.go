package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var greet string
	root := &cobra.Command{
		Use: "cli",
	}

	root.PersistentFlags().StringVar(
		&greet, "greet", "Hello", "Word to use as greetng")

	hello := &cobra.Command{
		Use:   "hello [name]",
		Short: "Greets user",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			uppercase, err := cmd.Flags().GetBool("uppercase")
			if err != nil {
				return err
			}
			if uppercase {
				name = strings.ToUpper(name)
			}
			fmt.Println(greet, name)
			return nil
		},
	}
	
	hello.Flags().BoolP("uppercase", "U", false, "Print name in uppercase")

	root.AddCommand(hello)

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
