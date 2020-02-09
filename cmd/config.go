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
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"os"
	"strconv"
	"strings"
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
		err := checkArgsNum(0, len(args), "")
	if err == nil {
		if printF {
			printConfig()
		}
		return
	}

	err = checkArgsNum(2, len(args), "")
	if err != nil {
		print(err.Error())
		return
	}

	changeConfig(args)
	},
}

func printConfig() {
	p, err := LoadProject()
	if err != nil {
		print(err.Error())
		return
	}
	b, err := p.ReadConfig()
	print(string(b))
}

func changeConfig(args []string) {
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

	switch strings.ToLower(args[0]) {
	case "name":
		config.Name = args[1]
	case "projectname":
		config.ProjectName = args[1]
	case "token":
		config.Token = args[1]
	case "serverdomain":
		config.ServerDomain = args[1]
	case "checkpoints":
		config.Checkpoints, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s is not an integer\n", args[1])
		}
	case "autopush":
		config.AutoPush = strings.ToLower(args[1]) == "true"
	default:
		fmt.Fprintf(os.Stderr, "%s is not a configuration field\n", args[0])
	}

	err = p.WriteConfig(config)
	if err != nil {
		print(err.Error())
		return
	}
}

func init() {
	configCmd.Flags().BoolVarP(&printF, "print", "p", false, "print configuration file")
	rootCmd.AddCommand(configCmd)
}
