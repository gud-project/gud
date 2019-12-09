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

	gudDir := path.Join(dir, ".gud")
	err := os.Mkdir(gudDir, os.ModeDir)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(path.Join(gudDir, "objects"), os.ModeDir)
	if err != nil {
		return nil, err
	}

	var f *os.File
	f, err = os.Create(path.Join(gudDir, "head"))
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	err = InitIndex(path.Join(gudDir, "index"))
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{gudDir}, nil
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

func (p *Project) Add(paths ...string) error {
	return AddToIndexFile(p.path, path.Join(p.path, ".gud/index"), paths)
}
