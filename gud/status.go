package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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

	root, err := loadTree(p.Path, version.TreeHash)
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

	// dont enter .gud
	if relPath == "." {
		ind := sort.Search(len(dir), func(i int) bool {
			return gudPath <= dir[i].Name()
		})
		if ind < len(dir) && dir[ind].Name() == gudPath {
			copy(dir[ind:], dir[ind+1:])
			dir = dir[:len(dir)-1]
		}
	}

	fileInd := 0
	objInd := 0
	for fileInd < len(dir) && objInd < len(root) {
		info := dir[fileInd]
		obj := root[objInd]
		basePath := info.Name()
		childPath := filepath.Join(relPath, basePath)

		if basePath < obj.Name { // new file/dir
			err = reportNew(rootPath, childPath, info.IsDir(), index, fn)
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
				err = reportNewDir(rootPath, childPath, index, fn)
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
		err = reportNew(rootPath, filepath.Join(relPath, info.Name()), info.IsDir(), index, fn)
	}
	for ; objInd < len(root); objInd++ {
		obj := root[objInd]
		err = reportRemoved(rootPath, filepath.Join(relPath, obj.Name), obj.Type == typeTree, obj.Hash, fn)
	}

	return nil
}

func reportNew(rootPath, relPath string, isDir bool, index []indexEntry, fn ChangeCallback) error {
	if isDir {
		return reportNewDir(rootPath, relPath, index, fn)
	}
	return reportNewFile(rootPath, relPath, index, fn)
}

func reportRemoved(rootPath, relPath string, isDir bool, hash objectHash, fn ChangeCallback) error {
	if isDir {
		return reportRemovedDir(rootPath, relPath, hash, fn)
	}
	return fn(relPath, StateRemoved)
}

func compareDir(rootPath, relPath string, hash objectHash, index []indexEntry, fn ChangeCallback) error {
	inner, err := loadTree(rootPath, hash)
	if err != nil {
		return err
	}

	return compareTree(rootPath, relPath, inner, index, fn)
}

func reportNewDir(rootPath, relPath string, index []indexEntry, fn ChangeCallback) error {
	return filepath.Walk(filepath.Join(rootPath, relPath), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			newRelPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			err = reportNewFile(rootPath, newRelPath, index, fn)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func reportNewFile(rootPath, relPath string, index []indexEntry, fn ChangeCallback) error {
	ind, tracked := findEntry(index, relPath)
	if tracked {
		entry := index[ind]
		if entry.State == StateNew || entry.State == StateModified {
			same, err := compareToObject(rootPath, relPath, entry.Hash)
			if err != nil {
				return err
			}
			if !same {
				return fn(relPath, StateModified)
			}
			return nil
		}
	}

	return fn(relPath, StateNew)
}

func reportRemovedDir(rootPath, relPath string, hash objectHash, fn ChangeCallback) error {
	tree, err := loadTree(rootPath, hash)
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
