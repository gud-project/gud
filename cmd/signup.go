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
	Long: `Create a new user on a remote server.
Server domain is in config file.`,
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

		var gConfig gud.GlobalConfig
		err = gud.LoadConfig(&gConfig, gConfig.GetPath())
		if err != nil {
			return err
		}

		resp, err := http.Post(fmt.Sprintf("%s/api/v1/signup", gConfig.ServerDomain), "application/json", &buf)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New("failed to create user")
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
	for err != nil && *password == "1" {
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
