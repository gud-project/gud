package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"net/http"
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

		var gConfig gud.GlobalConfig
		err = gud.LoadConfig(&gConfig, gConfig.GetPath())
		if err != nil {
			return err
		}

		return PullBranch(p, gConfig.ServerDomain)
	},
}

func PullBranch(p *gud.Project, domain string) (err error) {
	var config gud.Config
	err = p.LoadConfig(&config)
	if err != nil {
		return
	}

	var gConfig gud.GlobalConfig
	err = gud.LoadConfig(&gConfig, gConfig.GetPath())
	if err != nil {
		return
	}

	branch, err := p.CurrentBranch()
	if err != nil {
		return
	}
	hash, err := p.GetBranch(branch)
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/user/%s/project/%s/pull?branch=%s&start=%s", domain, config.OwnerName, config.ProjectName, branch, hash),
		nil)
	if err != nil {
		return
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: gConfig.Token})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = checkResponseError(resp)
	if err != nil {
		return err
	}

	_, err = p.PullBranch(branch, resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	return p.Reset()
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
