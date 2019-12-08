package gud

import (
	"os"
	"path"
	"path/filepath"
)

type Project struct {
	path string
}

func Start(dir string) (*Project, error) {
	// Check if got a path
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = wd
	}

	dir = path.Join(dir, ".gud")
	err := os.Mkdir(dir, os.ModeDir)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(path.Join(dir, "objects"), os.ModeDir)
	if err != nil {
		return nil, err
	}

	var f *os.File
	f, err = os.Create(path.Join(dir, "head"))
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{dir}, nil
}

func Load(dir string) (*Project, error) {
	for parent := filepath.Dir(dir); dir != parent; parent = filepath.Dir(parent) {
		info, err := os.Stat(path.Join(dir, ".gud"))
		if !os.IsNotExist(err) && info.IsDir() {
			return &Project{dir}, nil
		}
		dir = parent
	}

	return nil, Error{"No Gud project found at " + dir}
}
