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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}
		p, err := gud.Load(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}
		v, err := p.CurrentVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}

		_, err = p.Prev(*v)
		if err != nil && err.Error() == "The version has no predecessor" {
			fmt.Fprintf(os.Stdout, "Current branch has no versions\n")
		}

		err = printLog(*p, v)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	},
}

func printLog(p gud.Project, cv *gud.Version) error {
	prev, err := p.Prev(*cv)

	if err != nil {
		if err.Error() == "The version has no predecessor" {
			return nil
		}
		fmt.Fprintf(os.Stdout, cv.String())
		return err
	}
	if prev == nil {
		return nil
	}
	err = printLog(p, prev)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, cv.String())
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
}
