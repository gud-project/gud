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
		if len(args) > 1 {
			println("To many arguments in command call")
			return
		}
		if len(args) == 0 {
			println("Branch name required")
			return
		}

		token, err := loadToken()
		if err != nil {
			print(err)
			return
		}
		_ = token

		name := "name" // token.name
		pname := "pname" // token.pname
		data := "data" // token.data

		req, err := http.NewRequest("GET", fmt.Sprintf("localhost/api/v1/project/%s/%s/branch/%s", name, pname, args[0]), nil)
		if err != nil {
			println(err)
			return
		}
		req.AddCookie(&http.Cookie{Name: "session", Value: data})
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

		p , err:= LoadProject()
		if err != nil {
			println(err)
			return
		}

		var buf bytes.Buffer
		boundary, err := p.PushBranch(&buf, args[0], &hash)
		req, err = http.NewRequest("POST", fmt.Sprintf("localhost/api/v1/project/%s/%s/push?branch=%s", name, pname, args[0]), &buf)
		if err != nil {
			println(err)
			return
		}

		req.AddCookie(&http.Cookie{Name: "ds_user_id", Value: data})
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
