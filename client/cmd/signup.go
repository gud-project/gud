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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"golang.org/x/crypto/ssh/terminal"
	"net/http"
	"os"
	"regexp"
	"strings"
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
		fmt.Print("Enter user name: ")
		reader := bufio.NewReader(os.Stdin)
		name, err := reader.ReadString('\n')
		if err != nil {
			println(err.Error())
			return
		}

		email, _ := reader.ReadString('\n')
		for emailPattern.MatchString(email) {
			fmt.Fprintf(os.Stderr, "Use email format\n")
			email, _ = reader.ReadString('\n')
		}

		password, err := getPassword()
		for err != nil {
			println(err.Error())
			password, err = getPassword()
		}

		request := gud.SignUpRequest{name, email, password}

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
		err = saveToken(name, token)
	},
}

func getPassword() (string, error) {
	print("Enter password: ")
	p, err := getValidPassword()
	if err != nil {
		return "", err
	}
	print("Enter password again: ")
	vp, err := getValidPassword()
	if err != nil {
		return "", err
	}
	if vp != p {
		return "", errors.New("Passwords don't match\n")
	}

	return p, nil
}

func getValidPassword() (string, error) {
	bPassword, err := terminal.ReadPassword(0)
	if err != nil {
		return "", fmt.Errorf("Failed to read password: %s\n", err.Error())
	}
	password := string(bPassword)
	if len(password) < gud.PasswordLenMin {
		return "", errors.New("Password length must be more then 8 characters\n")
	}
	if strings.Contains(password, "&") {
		return "", errors.New("Password can't contain &\n")
	}
	return password, nil
}

func init() {
	rootCmd.AddCommand(signupCmd)


}
