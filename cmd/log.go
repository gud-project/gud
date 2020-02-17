package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show saved versions log",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}

		hash, err := p.CurrentHash()
		if err != nil {
			return err
		}
		v, err := p.CurrentVersion()
		if err != nil {
			return err
		}

		err = printLog(*p, *hash, *v)
		if err != nil {
			return err
		}

		return nil
	},
}

func printLog(p gud.Project, hash gud.ObjectHash, version gud.Version) error {
	if version.HasPrev() {
		prevHash, prev, err := p.Prev(version)
		if err != nil {
			return err
		}

		err = printLog(p, *prevHash, *prev)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stdout, "Message: %s\nTime: %s\nHash: %s\n\n",
		version.Message, version.Time.Format("2006-01-02 15:04:05"), hash)
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
}
