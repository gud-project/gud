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

	"github.com/spf13/cobra"
)

var stayF bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <branch name>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := checkArgsNum(1, len(args), "")
		if err != nil {
			print(err.Error())
			return
		}

		p, err := LoadProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load project: %s", err.Error())
			return
		}

		err = p.CreateBranch(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create branch: %s", err.Error())
			return
		}

		if !stayF {
			err = checkout(p, args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create branch: %s", err.Error())
			}
		}
	},
}

func init() {
	createCmd.Flags().BoolVarP(&stayF, "stay", "s", false, "Stay at the current branch")
	branchCmd.AddCommand(createCmd)
}
