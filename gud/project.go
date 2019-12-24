// gud is library containing the basics of a VCS
package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const gudPath = ".gud"
const headFileName = gudPath + "/head"

// Project is a representation of a Gud project
type Project struct {
	Path string
}

// Start creates a new Gud project in the path it receives.
// It returns a struct representing it.
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

	hash, err := initObjectsDir(dir)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, headFileName), hash[:], 0644)

	if err != nil {
		return nil, err
	}

	err = initIndex(dir)
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{dir}, nil
}

// Load receives a path to a Gud project and returns a representation of it.
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

// Add adds files to the current version of the Gud project
func (p *Project) Add(paths ...string) error {
	return addToIndex(p.Path, paths)
}

// Remove removes files from the current version of the Gud project
func (p *Project) Remove(paths ...string) error {
	return removeFromIndex(p.Path, paths)
}

// CurrentVersion returns the current version of the project
func (p Project) CurrentVersion() (*Version, error) {
	head, err := loadHead(p.Path)
	if err != nil {
		return nil, err
	}

	var currentVersion Version
	err = loadTree(p.Path, *head, &currentVersion)
	if err != nil {
		return nil, err
	}

	return &currentVersion, nil
}

// Save saves the current version of the project.
func (p *Project) Save(message string) (*Version, error) {
	index, err := loadIndex(p.Path)
	if err != nil {
		return nil, err
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return nil, err
	}

	var currentVersion Version
	err = loadTree(p.Path, *head, &currentVersion)
	if err != nil {
		return nil, err
	}

	dir := dirStructure{Name: "."}
	for _, entry := range index {
		addToStructure(&dir, entry.Name, entry.Hash)
	}

	var tree tree
	err = loadTree(p.Path, currentVersion.Tree, &tree)
	if err != nil {
		return nil, err
	}

	treeObj, err := buildTree(p.Path, "", dir, tree)
	if err != nil {
		return nil, err
	}

	newVersion := Version{
		Tree:    treeObj.Hash,
		Message: message,
		Time:    time.Now(),
		prev:    head,
	}

	versionObj, err := createVersion(p.Path, message, newVersion)
	if err != nil {
		return nil, err
	}

	err = writeHead(p.Path, versionObj.Hash)
	if err != nil {
		return nil, err
	}

	// reset index
	err = initIndex(p.Path)
	if err != nil {
		return nil, err
	}

	return &newVersion, nil
}

// Prev receives a version of the project and returns and it's previous one.
func (p Project) Prev(version Version) (*Version, error) {
	if !version.HasPrev() {
		return nil, Error{"The version has no predecessor"}
	}

	var prev Version
	err := loadTree(p.Path, *version.prev, &prev)
	if err != nil {
		return nil, err
	}

	return &prev, nil
}
