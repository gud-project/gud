package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		printStatus()
	},
}

func trackedCallback(relPath string, state gud.FileState) error {
	stateMsg := make(map[gud.FileState]string)
	stateMsg[gud.StateNew] = "new: "
	stateMsg[gud.StateRemoved] = "deleted: "
	stateMsg[gud.StateModified] = "modified: " //Change to empty when get a full message

	fMsg := stateMsg[state] + relPath + "\n"
	_, err := fmt.Fprintf(os.Stdout, fMsg)
	return err
}

func unTrackedCallback(relPath string, state gud.FileState) error {
	stateMsg := make(map[gud.FileState]string)
	stateMsg[gud.StateNew] = "non-update new:" //Change to empty when get a full message
	stateMsg[gud.StateRemoved] = "non-update deleted: "
	stateMsg[gud.StateModified] = "non-update modified: "

	fMsg := stateMsg[state] + relPath + "\n"
	_, err := fmt.Fprintf(os.Stdout, fMsg)
	return err
}

func printStatus() {
	p, err := LoadProject()
	if err != nil {
		return
	}

	err = p.Status(trackedCallback, unTrackedCallback)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
