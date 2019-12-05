package gud

import (
	"fmt"
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

	err = os.Mkdir(path.Join(dir, "/objects"), os.ModeDir)
	if err != nil {
		return nil, err
	}

	_, err = os.Create(dir)
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{dir}, nil
}

func Load(dir string) (*Project, error) {
	last := dir
	for parent := filepath.Dir(dir); last == parent; parent = filepath.Dir(parent) {

		info, err := os.Stat(path.Join(dir, ".gud"))
		if os.IsExist(err) && info.IsDir() {
			return &Project{dir}, nil
		}
		last = parent
	}

	return nil, fmt.Errorf("No Gud project found at " + dir)
}
