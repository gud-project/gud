package cmd

import (
	"errors"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gud",
	Short: "Gud is a version control system, that helps you manage your project's different versions",
	Long: `Gud is a version control system, attempting to answer the problems coming up with the
existing version controls systems, and add new, modern features to them.
Some of the new major features are an undo command in the CLI,
and automatic update. As for now, gud has one remote server working, in the automatically used domain "gud.codes"`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error { return errors.New("") },
}

var icons = survey.WithIcons(func(icons *survey.IconSet) {
	icons.Question.Text = ">>"
	icons.Question.Format = "blue+bh"
})

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
}
