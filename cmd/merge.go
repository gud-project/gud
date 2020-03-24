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

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <branch>",
	Short: "Merge the given branch into the current one",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkArgsNum(1, len(args), "")
		if err != nil {
			return err
		}

		p, err := LoadProject()
		if err != nil {
			return err
		}

		err = p.Checkpoint("merge")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		var dst gud.ObjectHash
		err = stringToHash(&dst, args[0])
		if err == nil {
			_, err = p.MergeHash(dst)
			if err != nil {
				_, err = mergeByName(p, args[0])
				if err != nil {
					return err
				}
			}

		} else {
			_, err = mergeByName(p, args[0])
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func mergeByName(p *gud.Project, name string) (v *gud.Version, err error) {
	v, err = p.MergeBranch(name)
	return
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
