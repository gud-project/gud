package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var recursF bool
var keepF bool

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm <file>...",
	Short: "rm receives the path of files in the project needed to be removed in the next save",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "missing files to remove")
		} else {
			if !recursF {
				var dirs []string
				for _, path := range args {
					file, err := os.Stat(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, err.Error())
						return
					}
					if mode := file.Mode(); mode.IsDir() {
						dirs = append(dirs, path)
					}
				}
				if len(dirs) > 0 {
					fmt.Fprintf(os.Stderr, "can not remove directories %s recursivle without -r", dirs)
				}
			}
			removeFiles(args)
		}
		if !keepF {
			deleteFiles(args)
		}
	},
}

func removeFiles(paths []string) {
	p, err := LoadProject()
	if err != nil {
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

	err = p.Remove(paths...)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func deleteFiles(paths []string) {
	for _, path := range paths {
		_ = os.RemoveAll(path)
	}
}

func init() {
	rmCmd.Flags().BoolVarP(&recursF, "recursive", "r", false, "remove directories recursively")
	rmCmd.Flags().BoolVarP(&keepF, "keep", "k", false, "keep the files after removing it from index, instead of deleting it")
	rootCmd.AddCommand(rmCmd)
}
