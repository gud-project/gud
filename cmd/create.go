package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var stayF bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "create <branch name>",
	Short: "Create a new branch",
	Long:  `A subcommand of "branch" root command. Used to create a new branch for a new version to work on`,
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]
		p, err := LoadProject()
		if err != nil {
			return fmt.Errorf("failed to load project: %s", err.Error())
		}

		err = p.Checkpoint("branch-create")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		err = p.CreateBranch(branchName)
		if err != nil {
			return err
		}

		if !stayF {
			err = checkout(p, branchName)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	createCmd.Flags().BoolVarP(&stayF, "stay", "s", false, "Stay at the current branch")
	branchCmd.AddCommand(createCmd)
}
