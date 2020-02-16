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
	"github.com/AlecAivazis/survey"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// checkoutCmd represents the checkout command
var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch>\ncheckout <commit-hash>",
	Short: "Transfer to another version of your project",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := LoadProject()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}

		if len(args) == 0 {
			for back := true; back ; {
				back, err = checkoutSelect(p)
				if err != nil {
					print(err.Error())
				}
			}
			return
		}

		err = checkArgsNum(1, len(args), "")
		if err != nil {
			print(err.Error())
			return
		}

		err = checkout(p, args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
		}
	},
}

func checkout(p *gud.Project, target string) error {
	var dst gud.ObjectHash
	err := stringToHash(&dst, target)
	if err != nil {
		err = p.CheckoutBranch(target)
		if err != nil {
			return err
		}
	}
	err = p.Checkout(dst)
	if err != nil {
		err = p.CheckoutBranch(target)
		if err != nil {
			return err
		}
	}
	return nil
}

func getVersionType() (string, error) {
	versionType := ""
	prompt := &survey.Select{
		Message: "Where do you want to checkout to:",
		Options: []string{"branch", "version"},
	}
	err := survey.AskOne(prompt, &versionType, icons)
	if err != nil {
		return "", err
	}
	return versionType, nil
}

func getBranch(p *gud.Project) (string, error) {
	var branches []string
	err := p.ListBranches(func(branch string) error {
		relBranch, _ := filepath.Rel(filepath.Join(p.Path, ".gud\\branches"), branch)
		branches = append(branches, relBranch)
		return nil
	})

	branches = append(branches, "back")
	branch := ""
	prompt := &survey.Select{
		Message: "Select a branch:",
		Options: branches,
	}

	err = survey.AskOne(prompt, &branch, icons)
	if err != nil {
		return "", err
	}

	return branch, nil
}

func getVersion(p *gud.Project, v *gud.Version) error {
	var versions []*gud.Version
	var versionMessages []string

	var prev *gud.Version
	var err error
	for v, err = p.CurrentVersion() ; err == nil && v.Message != "initial commit" ;  v = prev {
		versions = append(versions, v)
		versionMessages = append(versionMessages, v.Message)
		_, prev, err = p.Prev(*v)
	}

	if err != nil && err.Error() != "The version has no predecessor" {
		return errors.New("no versions created\n")
	}

	versionMessages = append(versionMessages, "back")
	versionMessage := ""
	prompt := &survey.Select{
		Message: "Select a version:",
		Options: versionMessages,
	}

	err = survey.AskOne(prompt, &versionMessage, icons)
	if err != nil {
		return err
	}

	if versionMessage == "back" {
		return errors.New("back")
	}

	for i := 0 ; i < len(versions) ; i++ {
		if versions[i].Message == versionMessage {
			v = versions[i]
			return nil
		}
	}

	return errors.New("version not found\n")
}
func checkoutSelect(p *gud.Project) (bool, error) {
	versionType, err := getVersionType()
	if err != nil {
		return false, err
	}

	switch versionType {
	case "branch":
		branch, err := getBranch(p)
		if err != nil {
			return false, err
		}

		if branch == "back" {
			return true, nil
		}

		err = p.CheckoutBranch(branch)
		if err != nil {
			return false, err
		}

	case "version":
		v, err := p.CurrentVersion()
		if err != nil {
			return false, err
		}

		_, _, err = p.Prev(*v)
		if err != nil {
			return false, err
		}

		var version gud.Version
		err = getVersion(p, &version)
		if err != nil {
			return err.Error() == "back", err
		}

		err = p.Checkout(version.TreeHash)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
