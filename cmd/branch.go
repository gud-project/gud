package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// branchCmd represents the branch command
var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Gives you information about your project's branches. Also takes place as the branch root command",
	Long: `branch is the root command for branch commands. This means in order to execute more
complex branch commands, you will write "gud branch" and then your command. In addition,
when branch is called by it's own it will print information about the branches in your project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}
		branch, err := p.CurrentBranch()
		fmt.Fprintf(os.Stdout, "Current branch is: \n%s\n", branch)
		fmt.Fprintf(os.Stdout, "Other branches:\n")
		err = p.ListBranches(func(branch string) error {
			_, err := fmt.Println(branch)
			return err
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
