// Package gud is library containing the basics of a VCS
package gud

import (
	"os"
	"path/filepath"
	"time"
)

const gudPath = ".gud"
const dirPerm = 0755

// Project is a representation of a Gud project
type Project struct {
	Path string
}

// Start creates a new Gud project in the path it receives.
// It returns a struct representing it.
func Start(dir string) (*Project, error) {
	project, err := StartHeadless(dir)
	if err != nil {
		return nil, err
	}

	tree, err := createTree(project.Path, "", tree{})
	if err != nil {
		return nil, err
	}

	obj, err := createVersion(project.Path, Version{
		Message:  initialCommitName,
		Time:     time.Now(),
		TreeHash: tree.Hash,
		prev:     nil,
	})
	if err != nil {
		return nil, err
	}

	err = dumpBranch(project.Path, FirstBranchName, obj.Hash)
	if err != nil {
		return nil, err
	}

	err = dumpHead(project.Path, Head{IsDetached: false, Branch: FirstBranchName})
	if err != nil {
		return nil, err
	}

	return project, err
}

func StartHeadless(dir string) (*Project, error) {
	// Check if got a path
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = wd
	}

	gudDir := filepath.Join(dir, gudPath)
	err := os.Mkdir(gudDir, dirPerm)
	if err != nil {
		return nil, err
	}

	err = initIndex(dir)
	if err != nil {
		return nil, err
	}

	err = initObjectsDir(dir)
	if err != nil {
		return nil, err
	}

	err = initBranches(dir)
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
func (p Project) Add(paths ...string) error {
	return addToIndex(p.Path, paths)
}

// Remove removes files from the current version of the Gud project
func (p Project) Remove(paths ...string) error {
	return removeFromProject(p.Path, paths)
}

// CurrentVersion returns the current version of the project
func (p Project) CurrentVersion() (*Version, error) {
	head, err := loadHead(p.Path)
	if err != nil {
		return nil, err
	}

	hash, err := getCurrentHash(p.Path, *head)
	if err != nil {
		return nil, err
	}

	return loadVersion(p.Path, *hash)
}

func (p Project) CurrentBranch() (string, error) {
	head, err := loadHead(p.Path)
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

	hash, err := loadBranch(p.Path, branch)
	if err != nil {
		return nil, err
	}

	return loadVersion(p.Path, *hash)
}

// Save saves the current version of the project.
func (p Project) Save(message string) (*Version, error) {
	index, err := loadIndex(p.Path)
	if err != nil {
		return nil, err
	}

	for _, entry := range index {
		if entry.State == StateConflict {
			return nil, Error{"conflicts must be solved before saving"}
		}
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return nil, err
	}

	if head.IsDetached {
		return nil, Error{"cannot save when head is detached"}
	}

	currentHash, err := getCurrentHash(p.Path, *head)
	if err != nil {
		return nil, err
	}

	currentVersion, err := loadVersion(p.Path, *currentHash)
	if err != nil {
		return nil, err
	}

	dir := dirStructure{Name: "."}
	for _, entry := range index {
		addToStructure(&dir, entry.Name, entry.Hash)
	}

	prev, err := loadTree(p.Path, currentVersion.TreeHash)
	if err != nil {
		return nil, err
	}

	treeObj, err := buildTree(p.Path, "", dir, prev)
	if err != nil {
		return nil, err
	}

	if treeObj == nil {
		treeObj, err = createTree(p.Path, message, tree{})
		if err != nil {
			return nil, err
		}
	}

	newVersion, err := saveVersion(p.Path, message, head.Branch, treeObj.Hash, currentHash, head.MergedHash)
	if err != nil {
		return nil, err
	}

	// reset index
	err = initIndex(p.Path)
	if err != nil {
		return nil, err
	}

	if head.MergedHash != nil {
		head.MergedHash = nil
		err = dumpHead(p.Path, *head)
		if err != nil {
			return nil, err
		}
	}

	return newVersion, nil
}

// Prev receives a version of the project and returns and it's previous one.
func (p Project) Prev(version Version) (*Version, error) {
	if !version.HasPrev() {
		return nil, Error{"The version has no predecessor"}
	}

	prev, err := loadVersion(p.Path, *version.prev)
	if err != nil {
		return nil, err
	}

	return prev, nil
}
