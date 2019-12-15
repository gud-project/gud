package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const gudPath = ".gud"
const headFileName = gudPath + "/head"

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

	err = ioutil.WriteFile(filepath.Join(gudDir, headFileName), hash[:], 0644)

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

func (p *Project) GetCurrentVersion() (*Version, error) {
	head, err := os.Open(filepath.Join(p.path, headFileName))
	if err != nil {
		return nil, err
	}

	var hash ObjectHash
	_, err = head.Read(hash[:])
	if err != nil {
		return nil, err
	}

	err = head.Close()
	if err != nil {
		return nil, err
	}

	var currentVersion Version
	err = LoadTree(p.path, hash, &currentVersion)

	return &currentVersion, nil
}

func (p *Project) Save(message string) error {
	//currentVersion, err := p.GetCurrentVersion()
	//if err != nil {
	//	return err
	//}

	index, err := loadIndex(filepath.Join(p.path, indexFilePath))
	if err != nil {
		return err
	}

	for _, entry := range index {
		for dir := filepath.Dir(entry.Name); dir != "."; dir = filepath.Dir(dir) {

		}
	}

	return nil
}
