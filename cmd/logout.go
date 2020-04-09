package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of a users",
	Long: `Logout from your connected gud user. Will effect every folder`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var config gud.GlobalConfig
		err := gud.LoadConfig(&config, config.GetPath())
		if err != nil {
			return err
		}

		config.Token = ""

		err = gud.WriteConfig(&config, config.GetPath())
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
