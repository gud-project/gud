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
	Short: "Prints the status of the current version compared to the last one",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printStatus()
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

func printStatus() error {
	p, err := LoadProject()
	if err != nil {
		return err
	}

	return p.Status(trackedCallback, unTrackedCallback)
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
