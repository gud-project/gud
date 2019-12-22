package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
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

	err = ioutil.WriteFile(filepath.Join(dir, headFileName), hash[:], 0644)

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
	head, err := LoadHead(p.path)
	if err != nil {
		return nil, err
	}

	var currentVersion Version
	err = LoadTree(p.path, *head, &currentVersion)
	if err != nil {
		return nil, err
	}

	return &currentVersion, nil
}

func (p *Project) Save(message string) (*Version, error) {
	index, err := loadIndex(filepath.Join(p.path, indexFilePath))
	if err != nil {
		return nil, err
	}

	head, err := LoadHead(p.path)
	if err != nil {
		return nil, err
	}

	var currentVersion Version
	err = LoadTree(p.path, *head, &currentVersion)
	if err != nil {
		return nil, err
	}

	dir := DirStructure{Name: "."}
	for _, entry := range index {
		AddToStructure(&dir, entry.Name, entry.Hash)
	}

	var tree Tree
	err = LoadTree(p.path, currentVersion.Tree, &tree)
	if err != nil {
		return nil, err
	}

	obj, err := BuildTree(p.path, "", dir, tree)
	if err != nil {
		return nil, err
	}

	newVersion := Version{
		Tree:    obj.Hash,
		Message: message,
		Time:    time.Now(),
		Prev:    head,
	}

	_, err = CreateVersion(p.path, message, newVersion)
	if err != nil {
		return nil, err
	}

	// reset index
	err = InitIndex(p.path)
	if err != nil {
		return nil, err
	}

	return &newVersion, nil
}
