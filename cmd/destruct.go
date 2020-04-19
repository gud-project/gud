package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"gitlab.com/magsh-2019/2/gud/gud"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var restartF bool

// destructCmd represents the destruct command
var destructCmd = &cobra.Command{
	Use:   "destruct",
	Short: "Delete .gud folder",
	Long: `Delete .gud folder,
therefor stopping track of file changes,
canceling Gud VCS and deleting saved old versions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}

		isSure := false
		prompt := &survey.Confirm{
			Message: "Are you sure you want to destruct? It will delete your .gud folder",
		}
		err = survey.AskOne(prompt, &isSure, icons)
		if err != nil {
			return err
		}

		if !isSure {
			return nil
		}

		_ = os.RemoveAll(filepath.Join(p.Path, gud.DefaultPath))
		if restartF {
			_, err = gud.Start("")
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	destructCmd.Flags().BoolVar(&restartF, "restart", false, "recreate the project")
	rootCmd.AddCommand(destructCmd)
}
