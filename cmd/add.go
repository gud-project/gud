package cmd

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		err := checkArgsNum(1, len(args), modeMin)
		if err != nil {
			if allF {
				files, err := getAllFiles()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get files: %s", err.Error())
				} else {
					addFiles(files)
				}
			} else {
				print(err)
			}
		} else {
			addFiles(args)
		}
	},
}

func addFiles(paths []string) {
	p, err := LoadProject()
	if err != nil {
		print(err)
		return
	}

	for i, path := range paths {
		temp, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can't use path %s: %s", path, err.Error())
			return
		}
		paths[i] = temp
	}

	err = p.Add(paths...)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func init() {
	addCmd.Flags().BoolVarP(&allF, "all", "a", false, "Add all files in the project")
	rootCmd.AddCommand(addCmd)
}
