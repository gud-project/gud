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
	"regexp"
	"strings"

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
	Run: func(cmd *cobra.Command, args []string) {
		print("Username: ")
		var name string
		fmt.Scanln(&name)

		print("Email: ")
		var email string
		fmt.Scanln(&email)
		for !emailPattern.MatchString(email) {
			print("Use email format\nEmail: ")
			fmt.Scanln(&email)
		}

		password, err := getPassword()
		for err != nil && password == "1"{
			print(err.Error())
			password, err = getPassword()
		}

		if err != nil && password == "2" {
			print(err.Error())
			return
		}

		request := gud.SignUpRequest{Username: name, Email: email, Password: password}

		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(request)
		if err != nil {
			println(err.Error())
			return
		}

		resp, err := http.Post("http://localhost/api/v1/signup", "application/json", &buf)
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

func getPassword() (string, error) {
	print("Password: ")
	p, err := getValidPassword()
	if err != nil {
		return p, err
	}
	print("password verification: ")
	vp, err := getValidPassword()
	if err != nil {
		return vp, err
	}
	if vp != p {
		return "1", errors.New("Passwords don't match\n")
	}

	return p, nil
}

func getValidPassword() (string, error) {
	var password string
	fmt.Scanln(&password)
	if len(password) < gud.PasswordLenMin {
		return "1", errors.New("Password length must be more then 8 characters\n")
	}
	if strings.Contains(password, "@") {
		return "1", errors.New("Password can't contain @\n")
	}
	return password, nil
}

func init() {
	rootCmd.AddCommand(signupCmd)
}
