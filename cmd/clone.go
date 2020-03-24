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
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone <URL>",
	Short: "Create a copy of a project in the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var domain string
		if err := checkArgsNum(0, len(args), modeMax) ; err == nil {
			var gConf gud.GlobalConfig
			err = gud.LoadConfig(&gConf, gConf.GetPath())
			if err != nil {
				return err
			}
			domain = gConf.ServerDomain
		} else {
			if len(args) == 1 {
				domain = args[0]
			} else {
				return err
			}
		}

		p, err := gud.StartHeadless("")
		if err != nil {
			return err
		}

		err = PullBranch(p, domain)
		if err != nil {
			return err
		}

		return p.AddHead()
	},
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
