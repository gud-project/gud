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
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+$`)

// signupCmd represents the signup command
var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Create a new user in the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name, email, password string
		err := getUserData(&name, &email, &password)
		if err != nil {
			return err
		}

		request := gud.SignUpRequest{Username: name, Email: email, Password: password}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(request)
		if err != nil {
			return err
		}

		resp, err := http.Post("http://localhost/api/v1/signup", "application/json", &buf)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		var token string
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "session" {
				token = cookie.Value
			}
		}

		var config gud.GlobalConfig
		err = gud.LoadConfig(&config, config.Token)
		if err != nil {
			return err
		}

		config.Name = name
		config.Token = token

		return gud.WriteConfig(&config, config.GetPath())
	},
}

func getUserData(name, email, password *string) error {
	prompt := &survey.Input{
		Message: "Username:",
	}
	err := survey.AskOne(prompt, name, icons)
	if err != nil {
		return err
	}

	prompt = &survey.Input{
		Message: "Email:",
	}
	err = survey.AskOne(prompt, email, icons)
	if err != nil {
		return err
	}

	for !emailPattern.MatchString(*email) {
		fmt.Fprintf(os.Stderr, "Use email format\n")
		err = survey.AskOne(prompt, email, icons)
		if err != nil {
			return err
		}
	}

	*password, err = getPassword()
	for err != nil && *password == "1"{
		fmt.Fprintf(os.Stderr, err.Error())
		*password, err = getPassword()
	}

	if err != nil && *password == "2" {
		return err
	}
	return nil
}

func getPassword() (string, error) {
	p, err := getValidPassword("Password:")
	if err != nil {
		return p, err
	}
	vp, err := getValidPassword("password verification:")
	if err != nil {
		return vp, err
	}
	if vp != p {
		return "1", errors.New("Passwords don't match\n")
	}

	return p, nil
}

func getValidPassword(message string) (string, error) {
	password := ""
	prompt := &survey.Password{
		Message: message,
	}
	err := survey.AskOne(prompt, &password, icons)
	if err != nil {
		return "", nil
	}
	if len(password) < gud.PasswordLenMin {
		return "1", fmt.Errorf("Password length must be %d characters or more\n", gud.PasswordLenMin)
	}
	if strings.Contains(password, "@") {
		return "1", errors.New("Password can't contain @\n")
	}
	return password, nil
}

func init() {
	rootCmd.AddCommand(signupCmd)
}
