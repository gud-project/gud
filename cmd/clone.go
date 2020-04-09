package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"net/http"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Args:  cobra.RangeArgs(2, 3),
	Use:   "clone [domain] <owner> <project>",
	Short: "Create a copy of a project in the server",
	Long: `Start a new project in the current folder. In difference from start, the project will not be empty,
but will be a copy of an existing project from the server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var domain, owner, project string
		if len(args) == 2 {
			var gConf gud.GlobalConfig
			err := gud.LoadConfig(&gConf, gConf.GetPath())
			if err != nil {
				return err
			}
			domain, owner, project = gConf.ServerDomain, args[0], args[1]
		} else {
			domain, owner, project = args[0], args[1], args[2]
		}

		p, err := gud.StartHeadless("")
		if err != nil {
			return err
		}

		var gConfig gud.GlobalConfig
		err = gud.LoadConfig(&gConfig, gConfig.GetPath())
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodGet,
			fmt.Sprintf("http://%s/api/v1/user/%s/project/%s/pull?branch=%s",
				domain, owner, project, gud.FirstBranchName), nil)
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
		err = checkResponseError(resp)
		if err != nil {
			return err
		}

		err = p.PullBranch(gud.FirstBranchName, resp.Body, resp.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		err = p.AddHead()
		if err != nil {
			return err
		}

		var config gud.Config
		err = p.LoadConfig(&config)
		if err != nil {
			return err
		}

		config.OwnerName = owner
		config.ProjectName = project
		err = p.WriteConfig(config)
		if err != nil {
			return err
		}

		return p.Reset()
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
