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
	Long: `Prints the status of all of the changes made from the last save.
Will show any changed removed or added file or folder`,
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
