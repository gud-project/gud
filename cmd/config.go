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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

var printF = false

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <configuration-key> <new-value>",
	Short: "Give a new value to a filed in the configuration file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := LoadProject()
		if err != nil {
			print(err.Error())
			return
		}

		var config gud.Config
		err = p.LoadConfig(&config)
		if err != nil {
			print(err.Error())
			return
		}

		err = checkArgsNum(0, len(args), "")
		if err == nil {
			if printF {
				err = printConfig(p)
				if err != nil {
					print(err)
				}
				return
			}
		}

		if len(args) != 2 && len(args) != 0 {
			print(err)
			return
		}

		err = getConfigChanges(args, &config)
		if err != nil {
			print(err.Error())
			return
		}

		err = p.WriteConfig(config)
		if err != nil {
			print(err.Error())
		}
	},
}

func printConfig(p *gud.Project) error{
	b, err := p.ReadConfig()
	print(string(b))
	return err
}

func getConfigChanges(args []string, config *gud.Config) error{
	var field, value string
	var err error
	if len(args) == 0 {
		prompt := &survey.Select{
			Message: "Choose field:",
			Options: []string{"Name", "Project name", "Token", "Server domain", "Checkpoints", "Automatic push"},
		}
		err = survey.AskOne(prompt, &field, icons)
		if err != nil {
			return err
		}

		if field == "Automatic push" {
			value := false
			prompt := &survey.Confirm{
				Message: "Do you want automatic push?",
			}
			err = survey.AskOne(prompt, &value, icons)
			if err != nil {
				return err
			}
			config.AutoPush = value
			return nil
		}

		newValue := &survey.Input{
			Message: "New value:",
		}
		err = survey.AskOne(newValue, &value, icons)
		if err != nil {
			return err
		}
	} else {
		field = args[0]
		value = args[1]
	}

	switch strings.ToLower(field) {
	case "name":
		config.Name = value
	case "projectname":
		config.ProjectName = value
	case "token":
		config.Token = value
	case "serverdomain":
		config.ServerDomain = value
	case "checkpoints":
		var err error
		config.Checkpoints, err = strconv.Atoi(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s is not an integer\n", value)
		}
	case "autopush":
		config.AutoPush = value == "true"
	default:
		fmt.Fprintf(os.Stderr, "%s is not a configuration field\n", field)
	}

	return nil
}

func init() {
	configCmd.Flags().BoolVarP(&printF, "print", "p", false, "print configuration file")
	rootCmd.AddCommand(configCmd)
}
