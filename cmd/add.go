package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var allF bool

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <file>...",
	Short: "Add receives the path of the updated files in the project, in order to use them in the next save",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkArgsNum(1, len(args), modeMin)
		if err != nil {
			if allF {
				files, err := getAllFiles()
				if err != nil {
					return err
				} else {
					return addFiles(files)
				}
			} else {
				return err
			}
		} else {
			return addFiles(args)
		}
	},
}

func addFiles(paths []string) error {
	p, err := LoadProject()
	if err != nil {
		return err
	}
	err = p.Checkpoint("add")
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = p.Undo()
		}
	}()

	for i, path := range paths {
		temp, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		paths[i] = temp
	}

	err = p.Add(paths...)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	addCmd.Flags().BoolVarP(&allF, "all", "a", false, "Add all files in the project")
	rootCmd.AddCommand(addCmd)
}
