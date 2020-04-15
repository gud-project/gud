package cmd

import (
	"github.com/spf13/cobra"
)

// undoCmd represents the undo command
var undoCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "undo",
	Short: "Undo the last command",
	Long: `Return to the version before the last command was executed.
Only commands that changes information counts.
Number of last versions saved can be modified using config file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
