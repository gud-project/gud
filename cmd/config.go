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
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

var printF = false
var globalF = false

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <configuration-key> <new-value>",
	Short: "Give a new value to a filed in the configuration file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var config gud.Config
		var gConfig gud.GlobalConfig

		var p *gud.Project
		var err error

		if globalF {
			err = gud.LoadConfig(gConfig, gConfig.GetPath())
		} else {
			p, err = LoadProject()
			if err != nil {
				return err
			}

			err = p.LoadConfig(&config)
			if err != nil {
				return err
			}
		}


		err = checkArgsNum(0, len(args), "")
		if err == nil {
			if printF {
				if globalF {
					err = printGlobalConfig(gConfig.GetPath())
				} else {
					err = printConfig(p)
				}
				if err != nil {
					return err
				}
				return nil
			}
		}

		if len(args) != 2 && len(args) != 0 {
			return err
		}

		if globalF {
			err = getGlobalConfigChanges(args, &gConfig)
			if err != nil {
				return err
			}
		} else {
			err = getConfigChanges(args, &config)
			if err != nil {
				return err
			}
		}

		err = p.Checkpoint("config-change")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		if globalF {
			err = gud.WriteConfig(gConfig, gConfig.GetPath())
			if err != nil {
				println(1)
				return err
			}
		} else {
			err = p.WriteConfig(config)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func printConfig(p *gud.Project) error {
	b, err := p.ReadConfig()
	print(string(b))
	return err
}

func printGlobalConfig(path string) error {
	b, err := gud.ReadConfig(path)
	print(string(b))
	return err
}

func getConfigChanges(args []string, config *gud.Config) error {
	var field, value string
	var err error
	if len(args) == 0 {
		prompt := &survey.Select{
			Message: "Choose field:",
			Options: []string{"Project name", "Checkpoints", "Automatic push"},
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
	case "project name", "projectname":
		config.ProjectName = value
	case "checkpoints":
		var err error
		config.Checkpoints, err = strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("%s is not an integer\n", value)
		}
	case "automatic push", "automaticpush":
		config.AutoPush = value == "true"
	default:
		return fmt.Errorf("%s is not a configuration field\n", field)
	}

	return nil
}

func getGlobalConfigChanges(args []string, config *gud.GlobalConfig) error {
	var field, value string
	var err error
	if len(args) == 0 {
		prompt := &survey.Select{
			Message: "Choose field:",
			Options: []string{"Name", "Token", "Domain server"},
		}
		err = survey.AskOne(prompt, &field, icons)
		if err != nil {
			return err
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
	case "token":
		config.Token = value
	case "server domain", "serverdomain":
		config.ServerDomain = value
	default:
		return fmt.Errorf("%s is not a configuration field\n", field)
	}

	return nil
}

func init() {
	configCmd.Flags().BoolVarP(&printF, "print", "p", false, "print configuration file")
	configCmd.Flags().BoolVarP(&globalF, "global", "g", false, "use global configuration")
	rootCmd.AddCommand(configCmd)
}
