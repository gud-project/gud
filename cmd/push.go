package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"net/http"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push [branch]",
	Short: "Push current branch to server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := checkArgsNum(1, len(args), "")
		if err != nil {
			print(err.Error())
			return
		}

		err = pushBranch(args[0])
		if err != nil {
			print(err.Error())
		}
	},
}

func pushBranch(branch string) error {
	p , err:= LoadProject()
	if err != nil {
		return err
	}

	var config gud.Config
	err = p.LoadConfig(&config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/project/%s/%s/branch/%s", config.ServerDomain, config.Name, config.ProjectName, branch), nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "session", Value: config.Token})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println("c1")
		println(config.Token)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errors.New("branch not found\n")
	}

	var hash gud.ObjectHash
	_, err = resp.Body.Read(hash[:])
	if err != nil {
		println("2")
		return err
	}

	var buf bytes.Buffer
	boundary, err := p.PushBranch(&buf, branch, &hash)
	req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/user/%s/project/%s/push?branch=%s", config.ServerDomain, config.Name, config.ProjectName, branch), &buf)
	if err != nil {
		println("3")
		return err
	}

	req.AddCookie(&http.Cookie{Name: "ds_user_id", Value: config.Token})
	req.Header.Add("Content-Type", "multipart/mixed; boundary=" +boundary)

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
