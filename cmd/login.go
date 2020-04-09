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
	Long: `Login to a gud user. Allows you to use clone, push and pull commands with our servers.
User will remain loged with every folder in your terminal until you logout. 
gud user can be created in our website at gud.codes`,
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
