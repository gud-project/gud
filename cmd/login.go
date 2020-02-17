/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login into a user's account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		prompt := &survey.Input{
			Message: "Username",
		}
		err := survey.AskOne(prompt, &name, icons)
		if err != nil {
			print(err.Error())
			return
		}

		password, err := getValidPassword("Password")
		if err != nil {
			println(err.Error())
			return
		}

		request := gud.LoginRequest{Username: name, Password: password, Remember: true}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(request)
		if err != nil {
			println(err.Error())
			return
		}

		resp, err := http.Post("http://localhost/api/v1/login", "application/json", &buf)
		if err != nil {
			println(err.Error())
			return
		}
		defer resp.Body.Close()

		var token string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session" {
				token = cookie.Value
			}
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

		config.Name = name
		config.Token = token

		err = p.WriteConfig(config)
		if err != nil {
			print(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
