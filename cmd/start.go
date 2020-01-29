package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [path]",
	Short: "Create a gud project in a given path",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Run: func(cmd *cobra.Command, args []string) {

		var err error
		if len(args) == 0 {
			_, err = gud.Start("")
		} else {
			_, err = gud.Start(args[0])
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
