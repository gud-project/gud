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
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
	"log"
	"os"
	"strings"
)

// checkoutCmd represents the checkout command
var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch>\ncheckout <commit-hash>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Required branch or commit hash to go to\n")
			return
		}
		p, err := LoadProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
		err = checkout(p, args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	},
}

func checkout(p *gud.Project, target string) error {
	err := p.CheckoutBranch(target)
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the path specified") {
			src := []byte(target)
			var dst gud.ObjectHash
			_, err := hex.Decode(dst[:], src)
			if err != nil {
				log.Fatal(err)
			}

			err = p.Checkout(dst)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
