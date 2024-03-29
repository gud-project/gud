package cmd

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gitlab.com/magsh-2019/2/gud/gud"
)

// checkoutCmd represents the checkout command
var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch>\ncheckout <commit-hash>",
	Short: "Transfer to another version of your project",
	Long: `Transfer to another saved version of your project.
Allows user to make changes in different versions of the project,
While the other versions are preserved`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := LoadProject()
		if err != nil {
			return err
		}

		err = p.Checkpoint("checkpoint")
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.Undo()
			}
		}()

		if len(args) == 0 {
			for back := true; back; {
				back, err = checkoutSelect(p)
				if err != nil {
					return err
				}
			}
			return nil
		}

		err = checkArgsNum(1, len(args), "")
		if err != nil {
			return err
		}

		err = checkout(p, args[0])
		if err != nil {
			return err
		}

		return nil
	},
}

func checkout(p *gud.Project, target string) error {
	var dst gud.ObjectHash
	err := stringToHash(&dst, target)
	if err != nil {
		return p.CheckoutBranch(target)
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
		branches = append(branches, branch)
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
	for v, err = p.CurrentVersion(); err == nil && v.Message != "initial commit"; v = prev {
		versions = append(versions, v)
		versionMessages = append(versionMessages, v.Message)
		_, prev, err = p.Prev(*v)
	}

	if err != nil && err.Error() != "The version has no predecessor" {
		return errors.New("no versions created")
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

	for i := 0; i < len(versions); i++ {
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
