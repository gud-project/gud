package cmd

import (
	"bytes"
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

		p , err:= LoadProject()
		if err != nil {
			print(err)
			return
		}

		var config gud.Config
		err = p.LoadConfig(&config)
		if err != nil {
			print(err)
			return
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/project/%s/%s/branch/%s", config.ServerDomain, config.Name, config.ProjectName, args[0]), nil)
		if err != nil {
			println(err)
			return
		}
		req.AddCookie(&http.Cookie{Name: "session", Value: config.Token})
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			println(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			print("branch not found")
			return
		}

		var hash gud.ObjectHash
		_, err = resp.Body.Read(hash[:])
		if err != nil {
			println(err)
			return
		}

		var buf bytes.Buffer
		boundary, err := p.PushBranch(&buf, args[0], &hash)
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/project/%s/%s/push?branch=%s", config.ServerDomain, config.Name, config.ProjectName, args[0]), &buf)
		if err != nil {
			println(err)
			return
		}

		req.AddCookie(&http.Cookie{Name: "ds_user_id", Value: config.Token})
		req.Header.Add("Content-Type", "multipart/mixed; boundary=" +boundary)

		resp, err = client.Do(req)
		if err != nil {
			println(err)
			return
		}
		resp.Body.Close()
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
