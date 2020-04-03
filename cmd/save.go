package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

var message string

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "save -m [message]",
	Short: "saves the current version of the project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if message == "" {
			prompt := &survey.Multiline{
				Message: "Enter commit message:",
			}
			err := survey.AskOne(prompt, &message, icons)
			if err != nil {
				return err
			}
		}

		p, err := LoadProject()
		if err != nil {
			return err
		}

		err = p.Checkpoint("save")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		_, err = saveVersion(p, message)
		if err != nil {
			return err
		}

		var config gud.Config
		err = p.LoadConfig(&config)
		if err != nil {
			return err
		}

		if config.AutoPush {
			err = pushBranch(message)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func saveVersion(p *gud.Project, message string) (*gud.Version, error) {
	v, err := p.Save(message)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func init() {
	saveCmd.Flags().StringVarP(&message, "message", "m", "", "Add all files in the project")
	rootCmd.AddCommand(saveCmd)
}
