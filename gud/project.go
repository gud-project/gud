// Package gud is library containing the basics of a VCS
package gud

import (
	"os"
	"path/filepath"
	"time"
)

const defaultGudPath = ".gud"
const dirPerm = 0755
const defaultCheckpointNum = 5

// Project is a representation of a Gud project
type Project struct {
	Path string
	gudPath string
}

// Start creates a new Gud project in the path it receives.
// It returns a struct representing it.
func Start(path string) (*Project, error) {
	project, err := startProject(path, defaultGudPath)
	if err != nil {
		return nil, err
	}

	_, err = startProject(project.Path, filepath.Join(defaultGudPath, defaultGudPath))
	if err != nil {
		return nil, err
	}

	return project, nil
}

func StartHeadless(dir string) (*Project, error) {
	return startGudDir(dir, defaultGudPath)
}

func startProject(path, gudRelPath string) (*Project, error) {
	project, err := startGudDir(path, gudRelPath)
	if err != nil {
		return nil, err
	}

	err = project.ConfigInit()
	if err != nil {
		return nil, err
	}

	tree, err := createTree(project.gudPath, "", tree{})
	if err != nil {
		return nil, err
	}

	obj, err := createVersion(project.gudPath, Version{
		Message:  initialCommitName,
		Author:   "Nitai", // TODO: get user from global config
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

	return project, nil
}

func startGudDir(path, gudRelPath string) (*Project, error) {
	// Check if got a path
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	gudPath := filepath.Join(abs, gudRelPath)
	err = os.Mkdir(gudPath, dirPerm)
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
	return &Project{abs, gudPath}, nil
}

// Load receives a path to a Gud project and returns a representation of it.
func Load(path string) (*Project, error) {
	for parent := filepath.Dir(path); path != parent; parent = filepath.Dir(parent) {
		gudPath := filepath.Join(path, defaultGudPath)
		info, err := os.Stat(gudPath)
		if !os.IsNotExist(err) && info.IsDir() {
			return &Project{path, gudPath}, nil
		}
		path = parent
	}

	return nil, Error{"No Gud project found at " + path}
}

func (p Project) CurrentHash() (*ObjectHash, error) {
	head, err := loadHead(p.gudPath)
	if err != nil {
		return nil, err
	}

	return getCurrentHash(p.gudPath, *head)
}

// CurrentVersion returns the current version of the project
func (p Project) CurrentVersion() (*Version, error) {
	hash, err := p.CurrentHash()
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
		addToStructure(&dir, entry.Path, entry.Hash)
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

	newVersion, err := p.saveVersion(message, head.Branch, treeObj.Hash, currentHash, head.MergedHash)
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

func (p Project) Checkpoint(message string) error {
	inner := p.innerProject()

	err := inner.AddAll()
	if err != nil {
		return err
	}

	version, err := inner.Save(message)
	if err != nil {
		return err
	}

	var lastHash, afterLastHash ObjectHash
	var afterLast Version
	last := *version
	i := 0
	for ; i < defaultCheckpointNum; i++ {
		if !last.HasPrev() {
			break
		}
		tmpHash, tmp, err := inner.Prev(last)
		if err != nil {
			return err
		}

		afterLast, afterLastHash = last, lastHash
		last, lastHash = *tmp, *tmpHash
	}

	if i == defaultCheckpointNum {
		err = removeVersion(p.gudPath, last, afterLast, lastHash, afterLastHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Project) Undo() error {
	inner := p.innerProject()

	err := inner.assertNoChanges()
	if err == ErrUnstagedChanges || err == ErrUnsavedChanges {
		return inner.Reset()
	}
	if err != nil {
		return err
	}

	head, err := loadHead(inner.gudPath)
	if err != nil {
		return err
	}
	hash, err := getCurrentHash(inner.gudPath, *head)
	if err != nil {
		return err
	}

	current, err := loadVersion(inner.gudPath, *hash)
	if err != nil {
		return err
	}

	if !current.HasPrev() {
		return Error{"nothing to undo"}
	}

	prevHash, _, err := inner.Prev(*current)
	if err != nil {
		return err
	}

	err = inner.Checkout(*prevHash)
	if err != nil {
		return err
	}

	tree, err := loadTree(inner.gudPath, current.TreeHash)
	if err != nil {
		return err
	}
	err = walkObjects(inner.gudPath, ".", tree, func(relPath string, obj object) error {
		return os.Remove(objectPath(inner.gudPath, obj.Hash))
	})
	err = os.Remove(objectPath(inner.gudPath, *hash))
	if err != nil {
		return err
	}

	err = dumpBranch(inner.gudPath, head.Branch, *prevHash)
	if err != nil {
		return err
	}

	return dumpHead(inner.gudPath, *head)
}

func (p Project) innerProject() Project {
	return Project{p.Path, filepath.Join(p.gudPath, defaultGudPath)}
}
