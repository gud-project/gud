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
	"errors"
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		prompt := &survey.Input{
			Message: "Username:",
		}
		err := survey.AskOne(prompt, &name, icons)
		if err != nil {
			return err
		}

		password, err := getValidPassword("Password:")
		if err != nil {
			return err
		}

		request := gud.LoginRequest{Username: name, Password: password, Remember: true}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(request)
		if err != nil {
			return err
		}

		var gConfig gud.GlobalConfig
		err = gud.LoadConfig(&gConfig, gConfig.GetPath())
		if err != nil {
			return err
		}

		resp, err := http.Post(fmt.Sprintf("http://%s/api/v1/login", gConfig.ServerDomain), "application/json", &buf)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New("incorrect username or password")
		}

		var token string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session" {
				token = cookie.Value
			}
		}

		var config gud.GlobalConfig
		err = gud.LoadConfig(&config, config.GetPath())
		if err != nil {
			return err
		}

		config.Name = name
		config.Token = token

		err = gud.WriteConfig(&config, config.GetPath())
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
