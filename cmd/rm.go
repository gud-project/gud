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
	Short: "remove given files from the project's version",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkArgsNum(1, len(args), modeMin)
		if err != nil {
			return err
		}

		if !recursF {
			var dirs []string
			for _, path := range args {
				file, err := os.Stat(path)
				if err != nil {
					return err
				}
				if mode := file.Mode(); mode.IsDir() {
					dirs = append(dirs, path)
				}
			}
			if len(dirs) > 0 {
				return fmt.Errorf("can not remove directories %s recursivle without -r", dirs)
			}
		}
		err = removeFiles(args)
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
	rmCmd.Flags().BoolVarP(&keepF, "keep", "k", false, "keep the files after removing it from index, instead of deleting it")
	rootCmd.AddCommand(rmCmd)
}
