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
	"github.com/AlecAivazis/survey"
	"gitlab.com/magsh-2019/2/gud/gud"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var restartF bool

// destructCmd represents the destruct command
var destructCmd = &cobra.Command{
	Use:   "destruct",
	Short: "Delete .gud folder",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}

		isSure := false
		prompt := &survey.Confirm{
			Message: "Are you sure you want to destruct? It will delete your .gud folder",
		}
		err = survey.AskOne(prompt, &isSure, icons)
		if err != nil {
			return err
		}

		if !isSure {
			return nil
		}

		isSure2 := false
		prompt = &survey.Confirm{
			Message: "Can you please not to? It really wants to live...",
		}
		err = survey.AskOne(prompt, &isSure2, icons)
		if err != nil {
			return err
		}

		if !isSure2{
			return nil
		}

		_ = os.RemoveAll(filepath.Join(p.Path, gud.DefaultPath))
		if restartF {
			_, err = gud.Start("")
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	destructCmd.Flags().BoolVar(&restartF, "restart", false, "recreate the project")
	rootCmd.AddCommand(destructCmd)
}
