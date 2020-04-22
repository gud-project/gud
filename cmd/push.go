package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"io"
	"net/http"
)

const projectNotFoundError = "project not found"

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Args:  cobra.MaximumNArgs(1),
	Use:   "push [branch]",
	Short: "Push current branch to server",
	Long: `Upload your branch into the server,
by creating or updating the branch on the server to be like your local one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return pushBranch("")
		}
		return pushBranch(args[0])
	},
}

func pushBranch(branch string) error {
	p, err := LoadProject()
	if err != nil {
		return err
	}

	if branch == "" {
		branch, err = p.CurrentBranch()
		if err != nil {
			return err
		}
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

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/user/%s/project/%s/branch/%s",
			gConfig.ServerDomain, config.OwnerName, config.ProjectName, branch), nil)
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
	if resp.StatusCode == http.StatusNotFound {
		var message gud.ErrorResponse
		err := json.NewDecoder(resp.Body).Decode(&message)
		if err != nil {
			return errors.New(resp.Status)
		}

		if message.Error == projectNotFoundError {
			err = createServerProject(config.ProjectName, gConfig)
			if err != nil {
				return err
			}
		}
	} else {
		err = checkResponseError(resp)
		if err != nil {
			return err
		}
	}

	var startHash *gud.ObjectHash
	if resp.StatusCode != http.StatusNotFound {
		var hash gud.ObjectHash
		_, err = resp.Body.Read(hash[:])
		if err != nil && err != io.EOF {
			return err
		}
		startHash = &hash
	}

	var buf bytes.Buffer
	boundary, err := p.PushBranch(&buf, branch, startHash)
	req, err = http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/v1/user/%s/project/%s/push?branch=%s",
			gConfig.ServerDomain, config.OwnerName, config.ProjectName, branch), &buf)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: gConfig.Token})
	req.Header.Add("Content-Type", "multipart/mixed; boundary="+boundary)

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = checkResponseError(resp)
	if err != nil {
		return err
	}

	return nil
}

func createServerProject(name string, gConf gud.GlobalConfig) error {
	request := gud.CreateProjectRequest{Name: name}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/api/v1/projects/import", gConf.ServerDomain), &buf)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{Name: "session", Value: gConf.Token})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	err = checkResponseError(resp)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
