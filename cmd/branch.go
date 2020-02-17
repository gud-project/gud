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

// branchCmd represents the branch command
var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Gives you information about your project's branches. Also takes place as the branch root command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := LoadProject()
		if err != nil {
			return
		}
		branch, err := p.CurrentBranch()
		fmt.Fprintf(os.Stdout, "Current branch is: \n%s\n", branch)
		fmt.Fprintf(os.Stdout, "Other branches:\n")
		err = p.ListBranches(func(branch string) error {
			_, err := fmt.Println(branch)
			return err
		})
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
