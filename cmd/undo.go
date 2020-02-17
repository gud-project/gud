package cmd

import (
	"github.com/spf13/cobra"
)

// undoCmd represents the undo command
var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkArgsNum(0, len(args), "")
		if err != nil {
			return err
		}

		p, err := LoadProject()
		if err != nil {
			return err
		}

		return p.Undo()
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
