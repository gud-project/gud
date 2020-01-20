package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

func getAllFiles() ([]string, error) {

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Checking if not adding .gud directory
		if !strings.Contains(path, ".gud") && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}
