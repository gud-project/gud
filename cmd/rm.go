package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var recursF bool
var keepF bool

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Args:  cobra.MinimumNArgs(1),
	Use:   "rm <file>...",
	Short: "Remove given files from the project's version",
	Long: `Remove a file from your current version of the project.
Will the delete the file(unless -k is used).
If the file was already deleted, it will only be removed on the index index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		//Check if there are directories to be removed without recursF
		if !recursF {
			var dirs []string
			for _, path := range args {
				file, err := os.Stat(path)
				if err != nil && strings.Contains(err.Error(), "The system cannot find the file specified.") {
					continue
				}

				if err != nil {
					return err
				}

				if mode := file.Mode(); mode.IsDir() {
					dirs = append(dirs, path)
				}
			}
			if len(dirs) > 0 {
				return fmt.Errorf("can not remove directories %s recursivly without -r", dirs)
			}
		}

		err := removeFiles(args)

		if !keepF {
			deleteFiles(args)
		}

		return err
	},
}

func removeFiles(paths []string) error {
	p, err := LoadProject()
	if err != nil {
		return err
	}

	err = p.Checkpoint("remove")
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
			return fmt.Errorf("can't use path %s: %s", path, err.Error())
		}
		paths[i] = temp
	}

	err = p.Remove(paths...)
	if err != nil {
		return err
	}

	return nil
}

func deleteFiles(paths []string) {
	for _, path := range paths {
		_ = os.RemoveAll(path)
	}
}

func init() {
	rmCmd.Flags().BoolVarP(&recursF, "recursive", "r", false, "remove directories recursively")
	rmCmd.Flags().BoolVarP(&keepF, "keep", "k", false, "keep the files after removing them from index, instead of deleting them")
	rootCmd.AddCommand(rmCmd)
}
