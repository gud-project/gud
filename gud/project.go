// Package gud is library containing the basics of a VCS
package gud

import (
	"os"
	"path/filepath"
	"time"
)

const defaultGudPath = ".gud"
const dirPerm = 0755

// Project is a representation of a Gud project
type Project struct {
	Path string
	gudPath string
}

// Start creates a new Gud project in the path it receives.
// It returns a struct representing it.
func Start(dir string) (*Project, error) {
	project, err := StartHeadless(dir)
	if err != nil {
		return nil, err
	}

	tree, err := createTree(project.gudPath, "", tree{})
	if err != nil {
		return nil, err
	}

	obj, err := createVersion(project.gudPath, Version{
		Message:  initialCommitName,
		Time:     time.Now(),
		TreeHash: tree.Hash,
	})
	if err != nil {
		return nil, err
	}

	err = dumpBranch(project.gudPath, FirstBranchName, obj.Hash)
	if err != nil {
		return nil, err
	}

	err = dumpHead(project.gudPath, Head{IsDetached: false, Branch: FirstBranchName})
	if err != nil {
		return nil, err
	}

	return project, err
}

func StartHeadless(dir string) (*Project, error) {
	return startGudDir(dir, defaultGudPath)
}

func startGudDir(path, gudRelPath string) (*Project, error) {
	// Check if got a path
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = wd
	}

	gudPath := filepath.Join(path, gudRelPath)
	err := os.Mkdir(gudPath, dirPerm)
	if err != nil {
		return nil, err
	}

	err = initIndex(gudPath)
	if err != nil {
		return nil, err
	}

	err = initObjectsDir(gudPath)
	if err != nil {
		return nil, err
	}

	err = initBranches(gudPath)
	if err != nil {
		return nil, err
	}

	// Create the directory
	return &Project{path, gudPath}, nil
}

// Load receives a path to a Gud project and returns a representation of it.
func Load(dir string) (*Project, error) {
	for parent := filepath.Dir(dir); dir != parent; parent = filepath.Dir(parent) {
		gudPath := filepath.Join(dir, defaultGudPath)
		info, err := os.Stat(gudPath)
		if !os.IsNotExist(err) && info.IsDir() {
			return &Project{dir, gudPath}, nil
		}
		dir = parent
	}

	return nil, Error{"No Gud project found at " + dir}
}

// CurrentVersion returns the current version of the project
func (p Project) CurrentVersion() (*Version, error) {
	head, err := loadHead(p.gudPath)
	if err != nil {
		return nil, err
	}

	hash, err := getCurrentHash(p.gudPath, *head)
	if err != nil {
		return nil, err
	}

	return loadVersion(p.gudPath, *hash)
}

func (p Project) CurrentBranch() (string, error) {
	head, err := loadHead(p.gudPath)
	if err != nil {
		return "", err
	}

	return head.Branch, nil
}

func (p Project) LatestVersion() (*Version, error) {
	branch, err := p.CurrentBranch()
	if err != nil {
		return nil, err
	}

	hash, err := loadBranch(p.gudPath, branch)
	if err != nil {
		return nil, err
	}

	return loadVersion(p.gudPath, *hash)
}

// Save saves the current version of the project.
func (p Project) Save(message string) (*Version, error) {
	index, err := loadIndex(p.gudPath)
	if err != nil {
		return nil, err
	}
	if len(index) == 0 {
		return nil, Error{"no changes to commit"}
	}
	for _, entry := range index {
		if entry.State == StateConflict {
			return nil, Error{"conflicts must be solved before saving"}
		}
	}

	head, err := loadHead(p.gudPath)
	if err != nil {
		return nil, err
	}

	if head.IsDetached {
		return nil, Error{"cannot save when head is detached"}
	}

	currentHash, err := getCurrentHash(p.gudPath, *head)
	if err != nil {
		return nil, err
	}

	currentVersion, err := loadVersion(p.gudPath, *currentHash)
	if err != nil {
		return nil, err
	}

	dir := dirStructure{Name: "."}
	for _, entry := range index {
		addToStructure(&dir, entry.Name, entry.Hash)
	}

	prev, err := loadTree(p.gudPath, currentVersion.TreeHash)
	if err != nil {
		return nil, err
	}

	treeObj, err := buildTree(p.gudPath, "", dir, prev)
	if err != nil {
		return nil, err
	}

	if treeObj == nil {
		treeObj, err = createTree(p.gudPath, message, tree{})
		if err != nil {
			return nil, err
		}
	}

	newVersion, err := saveVersion(p.gudPath, message, head.Branch, treeObj.Hash, currentHash, head.MergedHash)
	if err != nil {
		return nil, err
	}

	// reset index
	err = initIndex(p.gudPath)
	if err != nil {
		return nil, err
	}

	if head.MergedHash != nil {
		head.MergedHash = nil
		err = dumpHead(p.gudPath, *head)
		if err != nil {
			return nil, err
		}
	}

	return newVersion, nil
}

// Prev receives a version of the project and returns and it's previous one.
func (p Project) Prev(version Version) (*ObjectHash, *Version, error) {
	if !version.HasPrev() {
		return nil, nil, Error{"The version has no predecessor"}
	}

	prev, err := loadVersion(p.gudPath, *version.prev)
	if err != nil {
		return nil, nil, err
	}

	return version.prev, prev, nil
}
