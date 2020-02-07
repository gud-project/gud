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
	Run: func(cmd *cobra.Command, args []string) {
		err := checkArgsNum(0, len(args), "")
		if err != nil {
			print(err.Error())
			return
		}

		p, err := LoadProject()
		if err != nil {
			print(err.Error())
			return
		}

		err = p.Undo()
		if err != nil {
			print(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
