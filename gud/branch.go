package gud

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
)

const branchesDirPath = gudPath + "/branches"
const headFileName = gudPath + "/head"
const firstBranchName = "master"

type Head struct {
	IsDetached bool
	Branch     string
	Hash       objectHash
}

func (p Project) CreateBranch(name string) error {
	if name == "" {
		return Error{"branch name cannot be empty"}
	}

	if strings.ContainsRune(name, os.PathSeparator) {
		err := os.MkdirAll(filepath.Join(p.Path, branchesDirPath, filepath.Dir(name)), os.ModeDir)
		if err != nil {
			return err
		}
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return err
	}

	hash, err := getCurrentHash(p.Path, *head)
	if err != nil {
		return err
	}

	return dumpBranch(p.Path, name, *hash)
}

func (p Project) Checkout(hash objectHash) error {
	err := p.Status(
		func(relPath string, state FileState) error {
			return Error{"the index must be empty when checking out"}
		},
		func(relPath string, state FileState) error {
			return Error{"uncommitted changes must be cleaned before checking out"}
		},
	)
	if err != nil {
		return err
	}

	version, err := loadVersion(p.Path, hash)
	if err != nil {
		return err
	}
	tree, err := loadTree(p.Path, version.treeHash)
	if err != nil {
		return err
	}
	err = removeChanges(p.Path, tree)
	if err != nil {
		return err
	}

	head, err := loadHead(p.Path)
	if err != nil {
		return err
	}

	return dumpHead(p.Path, Head{
		IsDetached: true,
		Hash:       hash,
		Branch:     head.Branch,
	})
}

func removeChanges(rootPath string, tree tree) error {
	return compareTree(
		rootPath, ".", tree, []indexEntry{},
		func(relPath string, state FileState, hash *objectHash, isDir bool) error {
			path := filepath.Join(rootPath, relPath)
			if isDir && state == StateNew {
				return os.Remove(path)
			}
			if isDir {
				return os.Mkdir(path, os.ModeDir)
			}
			if state == StateNew {
				return os.Remove(path)
			}
			return unzipObject(rootPath, relPath, *hash)
		},
	)
}

func getCurrentHash(rootPath string, head Head) (*objectHash, error) {
	if head.IsDetached {
		return &head.Hash, nil
	}
	return loadBranch(rootPath, head.Branch)
}

func initBranches(rootPath string, firstHash objectHash) error {
	err := os.Mkdir(filepath.Join(rootPath, branchesDirPath), os.ModeDir)
	if err != nil {
		return err
	}

	err = dumpBranch(rootPath, firstBranchName, firstHash)
	if err != nil {
		return err
	}

	return dumpHead(rootPath, Head{IsDetached: false, Branch: firstBranchName})
}

func dumpBranch(rootPath string, name string, hash objectHash) (err error) {
	file, err := os.Create(filepath.Join(rootPath, branchesDirPath, name))
	if err != nil {
		return
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = file.Write(hash[:])
	return
}

func loadBranch(rootPath, name string) (*objectHash, error) {
	var hash objectHash

	file, err := os.Open(filepath.Join(rootPath, branchesDirPath, name))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	n, err := file.Read(hash[:])
	if err != nil {
		return nil, err
	}
	if n != len(hash) {
		return nil, Error{"branch is corrupted"}
	}

	return &hash, nil
}

func dumpHead(rootPath string, head Head) (err error) {
	file, err := os.Create(filepath.Join(rootPath, headFileName))
	if err != nil {
		return err
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	return gob.NewEncoder(file).Encode(head)
}

func loadHead(rootPath string) (*Head, error) {
	file, err := os.Open(filepath.Join(rootPath, headFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var head Head
	err = gob.NewDecoder(file).Decode(&head)
	if err != nil {
		return nil, err
	}

	return &head, nil
}
