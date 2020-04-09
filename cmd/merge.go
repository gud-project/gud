package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "merge <branch>",
	Short: "Merge the given branch into the current one",
	Long: ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}

		err = p.Checkpoint("merge")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		var dst gud.ObjectHash
		err = stringToHash(&dst, args[0])
		if err == nil {
			_, err = p.MergeHash(dst)
			if err != nil {
				_, err = mergeByName(p, args[0])
				if err != nil {
					return err
				}
			}

		} else {
			_, err = mergeByName(p, args[0])
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func mergeByName(p *gud.Project, name string) (v *gud.Version, err error) {
	v, err = p.MergeBranch(name)
	return
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
