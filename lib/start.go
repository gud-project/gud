package main

import (
	"os"
	"path"
)

func Start(dir string) error {
	// Check if got a path
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	// Create the directory
	return os.Mkdir(path.Join(dir, ".gud"), os.ModeDir)
}
