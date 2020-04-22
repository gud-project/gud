package cmd

import (
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "reset",
	Short: "Reset resets the project's state to the latest version",
	Long: `Reset removes all unstaged changes in the project,
returning it to the latest state that was saved.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}
		return p.Reset()
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
