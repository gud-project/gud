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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var restartF bool

// destructCmd represents the destruct command
var destructCmd = &cobra.Command{
	Use:   "destruct",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}
		root, err := getRoot(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
		_ = os.RemoveAll(root)
		if restartF {
			startCmd.Run(cmd, args)
		}
	},
}

func getRoot(path string) (string, error) {
	for parent := filepath.Dir(path); path != parent; parent = filepath.Dir(parent) {
		info, err := os.Stat(filepath.Join(path, ".gud"))
		if !os.IsNotExist(err) && info.IsDir() {
			return filepath.Join(path, ".gud"), nil
		}
		path = parent
	}
	return "", errors.New("No Gud project found\n")
}

func init() {
	destructCmd.Flags().BoolVar(&restartF, "restart", false, "recreate the project")
	rootCmd.AddCommand(destructCmd)
}
