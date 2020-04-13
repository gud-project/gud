package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"net/http"
)

const projectNotFound = ""

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push [branch]",
	Short: "Push current branch to server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkArgsNum(1, len(args), "")
		if err != nil {
			return err
		}

		return pushBranch(args[0])
	},
}

func pushBranch(branch string) error {
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
	err = gud.LoadConfig(&gConfig, gConfig.GetPath())
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/project/%s/%s/branch/%s", gConfig.ServerDomain, config.OwnerName, config.ProjectName, branch), nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: gConfig.Token})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if err.Error() == projectNotFound {
			err = createServerProject(config.ProjectName, gConfig.Token)
			if err != nil {
				return err
			}
			resp, err = client.Do(req)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer resp.Body.Close()

	var startHash *gud.ObjectHash
	if resp.StatusCode != http.StatusNotFound {
		var hash gud.ObjectHash
		_, err = resp.Body.Read(hash[:])
		if err != nil {
			return err
		}
		startHash = &hash
	}

	var buf bytes.Buffer
	boundary, err := p.PushBranch(&buf, branch, startHash)
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/project/%s/%s/push?branch=%s", gConfig.ServerDomain, gConfig.Name, config.ProjectName, branch), &buf)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "ds_user_id", Value: gConfig.Token})
	req.Header.Add("Content-Type", "multipart/mixed; boundary="+boundary)

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func createServerProject(name, token string) error {
	request := gud.CreateProjectRequest{Name: name}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost/api/v1/create"), &buf)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: token})

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
