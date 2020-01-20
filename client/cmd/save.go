package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

var message string

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save -m [message]",
	Short: "saves the current version of the project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "warning: no arguments required\n")
		}

		if message == "" {
			fmt.Fprintf(os.Stderr, "version message required. use -m\n")
		} else {
			saveVersion(message)
		}
	},
}

func saveVersion(message string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	p, err := gud.Load(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	_, err = p.Save(message)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func init() {
	saveCmd.Flags().StringVarP(&message, "message", "m", "", "Add all files in the project")
	rootCmd.AddCommand(saveCmd)
}
