package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const gudPath = ".gud"

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

	gudDir := filepath.Join(dir, gudPath)
	err := os.Mkdir(gudDir, os.ModeDir)
	if err != nil {
		return nil, err
	}

	hash, err := InitObjectsDir(dir)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filepath.Join(gudDir, "head"), hash[:], 0644)

	if err != nil {
		return nil, err
	}

	err = InitIndex(dir)
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{gudDir}, nil
}

func Load(dir string) (*Project, error) {
	for parent := filepath.Dir(dir); dir != parent; parent = filepath.Dir(parent) {
		info, err := os.Stat(filepath.Join(dir, ".gud"))
		if !os.IsNotExist(err) && info.IsDir() {
			return &Project{dir}, nil
		}
		dir = parent
	}

	return nil, Error{"No Gud project found at " + dir}
}

func (p *Project) Add(paths ...string) error {
	return AddToIndex(p.path, paths)
}

func (p *Project) Remove(paths ...string) error {
	return RemoveFromIndex(p.path, paths)
}

func (p *Project) Save(message string) error {
	return nil
}
