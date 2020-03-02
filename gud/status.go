package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type ChangeCallback func(relPath string, state FileState) error
type cmpCallback func(relPath string, state FileState, hash *ObjectHash, isDir bool) error

func (p Project) Status(trackedFn, untrackedFn ChangeCallback) error {
	index, err := p.getIndex()
	if err != nil {
		return err
	}

	for _, entry := range index {
		err = trackedFn(entry.Path, entry.State)
		if err != nil {
			return err
		}
	}

	version, err := p.CurrentVersion()
	if err != nil {
		return err
	}

	root, err := loadTree(p.gudPath, version.TreeHash)
	if err != nil {
		return err
	}

	return p.compareTree(".", root, index,
		func(relPath string, state FileState, hash *ObjectHash, isDir bool) error {
			return untrackedFn(relPath, state)
		},
	)
}

func (p Project) compareTree(relPath string, root tree, index []indexEntry, fn cmpCallback) error {
	dir, err := ioutil.ReadDir(filepath.Join(p.Path, relPath))
	if err != nil {
		return err
	}

	// dont enter .gud
	relGudPath, _ := filepath.Rel(p.Path, p.gudPath)
	if relPath == filepath.Dir(relGudPath) {
		gudBasePath := filepath.Base(relGudPath)
		ind := sort.Search(len(dir), func(i int) bool {
			return gudBasePath <= dir[i].Name()
		})
		if ind < len(dir) && dir[ind].Name() == gudBasePath {
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
			err = p.reportNew(childPath, info.IsDir(), index, fn)
			if err != nil {
				return err
			}

			fileInd++
		} else if obj.Name < basePath { // removed file/dir
			err = reportRemoved(p.gudPath, relPath, obj, index, fn)
			if err != nil {
				return err
			}

			objInd++
		} else {
			if obj.Type == typeBlob && info.IsDir() { // removed file and added directory
				err = fn(relPath, StateRemoved, &obj.Hash, false)
				if err != nil {
					return err
				}
				err = p.reportNewDir(childPath, index, fn)
				if err != nil {
					return err
				}

			} else if obj.Type == typeTree && !info.IsDir() { // removed directory and added file
				err = reportRemovedDir(p.gudPath, childPath, obj.Hash, index, fn)
				if err != nil {
					return err
				}
				err = fn(relPath, StateNew, nil, false)
				if err != nil {
					return err
				}

			} else if info.IsDir() {
				err = p.compareDir(childPath, obj.Hash, index, fn)
				if err != nil {
					return err
				}

			} else {
				err = p.compareFile(childPath, obj.Hash, index, fn)
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
		err = p.reportNew(filepath.Join(relPath, info.Name()), info.IsDir(), index, fn)
	}
	for ; objInd < len(root); objInd++ {
		obj := root[objInd]
		err = reportRemoved(p.gudPath, relPath, obj, index, fn)
	}

	return nil
}

func (p Project) reportNew(relPath string, isDir bool, index []indexEntry, fn cmpCallback) error {
	if isDir {
		return p.reportNewDir(relPath, index, fn)
	}
	return p.reportNewFile(relPath, index, fn)
}

func reportRemoved(gudPath, parentPath string, obj object, index []indexEntry, fn cmpCallback) error {
	relPath := filepath.Join(parentPath, obj.Name)
	if obj.Type == typeTree {
		return reportRemovedDir(gudPath, relPath, obj.Hash, index, fn)
	}
	return reportRemovedFile(relPath, obj.Hash, index, fn)
}

func (p Project) compareDir(relPath string, hash ObjectHash, index []indexEntry, fn cmpCallback) error {
	inner, err := loadTree(p.gudPath, hash)
	if err != nil {
		return err
	}

	return p.compareTree(relPath, inner, index, fn)
}

func (p Project) reportNewDir(relPath string, index []indexEntry, fn cmpCallback) error {
	return filepath.Walk(filepath.Join(p.Path, relPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		newRelPath, err := filepath.Rel(p.Path, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return fn(newRelPath, StateNew, nil, true)
		}
		return p.reportNewFile(newRelPath, index, fn)
	})
}

func (p Project) reportNewFile(relPath string, index []indexEntry, fn cmpCallback) error {
	ind, tracked := findEntry(index, relPath)
	if tracked {
		entry := index[ind]
		if entry.State == StateNew || entry.State == StateModified {
			same, err := p.compareToObject(relPath, entry.Hash)
			if err != nil {
				return err
			}
			if !same {
				return fn(relPath, StateModified, &entry.Hash, false)
			}
			return nil
		}
	}

	return fn(relPath, StateNew, nil, false)
}

func reportRemovedFile(relPath string, hash ObjectHash, index []indexEntry, fn cmpCallback) error {
	ind, tracked := findEntry(index, relPath)
	if !tracked || index[ind].State != StateRemoved {
		return fn(relPath, StateRemoved, &hash, false)
	}

	return nil
}

func reportRemovedDir(gudPath, relPath string, hash ObjectHash, index []indexEntry, fn cmpCallback) error {
	tree, err := loadTree(gudPath, hash)
	if err != nil {
		return err
	}

	err = walkObjects(gudPath, relPath, tree, func(relPath string, obj object) error {
		if obj.Type == typeBlob {
			return reportRemovedFile(relPath, obj.Hash, index, fn)
		}
		return fn(relPath, StateRemoved, &obj.Hash, true)
	})
	if err != nil {
		return err
	}

	return fn(relPath, StateRemoved, &hash, true)
}

func (p Project) compareFile(relPath string, hash ObjectHash, index []indexEntry, fn cmpCallback) error {
	ind, tracked := findEntry(index, relPath)
	if tracked {
		entry := index[ind]
		if entry.State == StateRemoved { // file was deleted and then added
			return fn(relPath, StateNew, nil, false)
		}

		hash = entry.Hash
	}

	same, err := p.compareToObject(relPath, hash)
	if err != nil {
		return err
	}
	if !same {
		err = fn(relPath, StateModified, &hash, false)
		if err != nil {
			return err
		}
	}

	return nil
}
