package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type ChangeCallback func(relPath string, state FileState) error

func (p Project) Status(trackedFn, untrackedFn ChangeCallback) error {
	index, err := loadIndex(p.Path)
	if err != nil {
		return err
	}

	for _, entry := range index {
		err = trackedFn(entry.Name, entry.State)
		if err != nil {
			return err
		}
	}

	version, err := p.CurrentVersion()
	if err != nil {
		return err
	}

	var root tree
	err = loadTree(p.Path, version.Tree, &root)
	if err != nil {
		return err
	}

	return compareTree(p.Path, ".", root, index, untrackedFn)
}

func compareTree(rootPath, relPath string, root tree, index []indexEntry, fn ChangeCallback) error {
	dir, err := ioutil.ReadDir(filepath.Join(rootPath, relPath))
	if err != nil {
		return err
	}

	fileInd := 0
	objInd := 0
	for fileInd < len(dir) && objInd < len(root) {
		info := dir[fileInd]
		obj := root[objInd]
		basePath := info.Name()
		childPath := filepath.Join(relPath, basePath)

		if basePath < obj.Name { // new file/dir
			err = reportNew(rootPath, childPath, info.IsDir(), fn)
			if err != nil {
				return err
			}

			fileInd++
		} else if obj.Name < basePath { // removed file/dir
			err = reportRemoved(rootPath, filepath.Join(relPath, obj.Name), obj.Type == typeTree, obj.Hash, fn)
			if err != nil {
				return err
			}

			objInd++
		} else {
			if obj.Type == typeBlob && info.IsDir() { // removed file and added directory
				err = fn(relPath, StateRemoved)
				if err != nil {
					return err
				}
				err = reportNewDir(rootPath, childPath, fn)
				if err != nil {
					return err
				}

			} else if obj.Type == typeTree && !info.IsDir() { // removed directory and added file
				err = reportRemovedDir(rootPath, childPath, obj.Hash, fn)
				if err != nil {
					return err
				}
				err = fn(relPath, StateNew)
				if err != nil {
					return err
				}

			} else if info.IsDir() {
				err = compareDir(rootPath, childPath, obj.Hash, index, fn)
				if err != nil {
					return err
				}

			} else {
				err = compareFile(rootPath, childPath, obj.Hash, index, fn)
				if err != nil {
					return err
				}
			}

			fileInd++
			objInd++
		}
	}

	for ; fileInd < len(dir); fileInd++ {
		info := dir[fileInd]
		err = reportNew(rootPath, filepath.Join(relPath, info.Name()), info.IsDir(), fn)
	}
	for ; objInd < len(root); objInd++ {
		info := dir[fileInd]
		err = reportNew(rootPath, filepath.Join(relPath, info.Name()), info.IsDir(), fn)
	}

	return nil
}

func reportNew(rootPath, relPath string, isDir bool, fn ChangeCallback) error {
	if isDir {
		return reportNewDir(rootPath, relPath, fn)
	}
	return fn(relPath, StateNew)
}

func reportRemoved(rootPath, relPath string, isDir bool, hash objectHash, fn ChangeCallback) error {
	if isDir {
		return reportRemovedDir(rootPath, relPath, hash, fn)
	}
	return fn(relPath, StateRemoved)
}

func compareDir(rootPath, relPath string, hash objectHash, index []indexEntry, fn ChangeCallback) error {
	var inner tree
	err := loadTree(rootPath, hash, &inner)
	if err != nil {
		return err
	}

	return compareTree(rootPath, relPath, inner, index, fn)
}

func reportNewDir(rootPath, relPath string, fn ChangeCallback) error {
	if relPath == gudPath { // don't enter .gud/
		return nil
	}
	return filepath.Walk(filepath.Join(rootPath, relPath), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			newRelPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			err = fn(newRelPath, StateNew)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func reportRemovedDir(rootPath, relPath string, hash objectHash, fn ChangeCallback) error {
	var tree tree
	err := loadTree(rootPath, hash, &tree)
	if err != nil {
		return err
	}

	return walkBlobs(rootPath, relPath, tree, func(relPath string) error {
		return fn(relPath, StateRemoved)
	})
}

func compareFile(rootPath, relPath string, hash objectHash, index []indexEntry, fn ChangeCallback) error {
	ind, tracked := findEntry(index, relPath)
	if tracked {
		entry := index[ind]
		if entry.State == StateRemoved { // file was deleted and then added
			return fn(relPath, StateNew)
		}

		hash = entry.Hash
	}

	same, err := compareToObject(rootPath, relPath, hash)
	if err != nil {
		return err
	}
	if !same {
		err = fn(relPath, StateModified)
		if err != nil {
			return err
		}
	}

	return nil
}
