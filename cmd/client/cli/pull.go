package main 


func NewPullCmd() *cobra.Command {

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
}
