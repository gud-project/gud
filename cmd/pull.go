package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [url]",
	Short: "Get the server's version of your branch",
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

		var config gud.Config
		err = p.LoadConfig(&config)
		if err != nil {
			return err
		}

		var gConfig gud.GlobalConfig
		err = gud.LoadConfig(gConfig, gConfig.GetPath())
		if err != nil {
			return err
		}

		branch, err := p.CurrentBranch()
		if err != nil {
			return err
		}
		hash, err := p.GetBranch(branch)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET",
			fmt.Sprintf("http://%s/api/v1/project/%s/%s/pull?branch=%s&start=%s", gConfig.ServerDomain, gConfig.Name, config.ProjectName, branch, hash),
			nil)
		if err != nil {
			return err
		}

		req.AddCookie(&http.Cookie{Name: "session", Value: gConfig.Token})
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return p.PullBranch(branch, resp.Body, resp.Header.Get("Content-Type"))
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
