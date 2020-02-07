package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		err := checkArgsNum(1, len(args), "")
		if err != nil {
			print(err.Error())
			return
		}

		name := "name" // token.name
		pname := "pname" // token.pname
		data := "data" // token.data

		p , err:= LoadProject()
		if err != nil {
			println(err)
			return
		}

		branch, err := p.CurrentBranch()
		if err != nil {
			println(err)
			return
		}
		hash, err := p.GetBranch(branch)
		if err != nil {
			println(err)
			return
		}

		req, err := http.NewRequest("GET",
			fmt.Sprintf("localhost/api/v1/project/%s/%s/pull?branch=%s&start=%s", name, pname, branch, hash),
			nil)
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

		err = p.PullBranch(branch, resp.Body, resp.Header.Get("Content-Type"))
		if err != nil {
			println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
