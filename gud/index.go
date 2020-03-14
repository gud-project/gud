package gud

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const indexFilePath = "index"

type FileState int

const (
	StateNew FileState = iota
	StateRemoved
	StateModified
	StateConflict
)

type indexEntry struct {
	Path  string
	Hash  ObjectHash
	State FileState
	Mtime time.Time
	Size  int64
}

type indexFile struct {
	Version PackageVersion
	Entries []indexEntry
}

// Add adds files to the current version of the Gud project
func (p Project) Add(paths ...string) error {
	// TODO: handle renames
	entries, err := loadIndex(p.gudPath)
	if err != nil {
		return err
	}

	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(p.Path, abs)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		prev, err := p.findObject(rel)
		if err != nil {
			return err
		}

		if info.IsDir() {
			var prevTree tree
			if prev != nil && prev.Type == typeTree {
				prevTree, err = loadTree(p.gudPath, prev.Hash)
				if err != nil {
					return err
				}
			}

			err = p.compareTree(
				rel, prevTree, entries, // TODO: might need to replace entries with nil
				func(relPath string, state FileState, hash *ObjectHash, isDir bool) error {
					if !isDir {
						entries, err = p.addIndexEntry(relPath, state, entries)
						if err != nil {
							return err
						}
					}
					return nil
				})
			if err != nil {
				return err
			}
		} else {
			var state FileState
			if prev != nil && prev.Type != typeTree {
				unchanged, err := p.compareToObject(rel, prev.Hash)
				if err != nil {
					return err
				}
				if unchanged {
					ind, found := findEntry(entries, rel)
					if found {
						copy(entries[ind:], entries[ind+1:])
						entries = entries[:len(entries)-1]
					}
				}
				state = StateModified
			} else {
				if prev != nil && prev.Type == typeTree {
					entries, err = p.removeDirFromIndex(rel, prev.Hash, entries)
					if err != nil {
						return err
					}
				}
				state = StateNew
			}

			entries, err = p.addIndexEntry(rel, state, entries)
			if err != nil {
				return err
			}
		}
	}

	return dumpIndex(p.gudPath, entries)
}

func (p Project) AddAll() error {
	return p.Add(p.Path)
}

// Remove removes files from the current version of the Gud project
func (p Project) Remove(paths ...string) error {
	entries, err := loadIndex(p.gudPath)
	if err != nil {
		return err
	}

	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(p.Path, abs)
		if err != nil {
			return err
		}

		prev, err := p.findObject(rel)
		if err != nil {
			return err
		}
		if prev == nil {
			if _, found := findEntry(entries, rel); !found {
				return Error{"untracked file: " + path}
			}
			entries, err = p.addIndexEntry(rel, StateRemoved, entries)
			if err != nil {
				return err
			}
		} else {
			if prev.Type == typeTree {
				entries, err = p.removeDirFromIndex(rel, prev.Hash, entries)
				if err != nil {
					return err
				}
			} else {
				entries, err = p.addIndexEntry(rel, StateRemoved, entries)
				if err != nil {
					return err
				}
			}
		}
	}

	return dumpIndex(p.gudPath, entries)
}

func (p Project) removeDirFromIndex(relPath string, prevHash ObjectHash, index []indexEntry) ([]indexEntry, error) {
	prevTree, err := loadTree(p.gudPath, prevHash)
	if err != nil {
		return nil, err
	}

	err = walkObjects(p.gudPath, relPath, prevTree, func(relPath string, obj object) error {
		if obj.Type != typeTree {
			index, err = p.addIndexEntry(relPath, StateRemoved, index)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return index, err
}

func (p Project) addIndexEntry(relPath string, state FileState, index []indexEntry) ([]indexEntry, error) {
	var mtime time.Time
	var n int64
	if state != StateRemoved {
		info, err := os.Stat(filepath.Join(p.Path, relPath))
		if err != nil {
			return nil, err
		}

		mtime = info.ModTime()
		n = info.Size()
	}

	ind, found := findEntry(index, relPath)
	if found {
		prevEntry := index[ind]
		if prevEntry.State != StateRemoved {
			if state == StateRemoved {
				err := removeEntry(p.gudPath, prevEntry.Hash)
				if err != nil {
					return nil, err
				}
				copy(index[ind:], index[ind+1:])
				return index[:len(index)-1], nil
			}

			if prevEntry.Mtime.Before(mtime) {
				unchanged, err := p.compareToObject(relPath, prevEntry.Hash)
				if err != nil {
					return nil, err
				}
				if unchanged {
					return index, nil
				}
				err = removeEntry(p.gudPath, prevEntry.Hash)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		index = append(index, indexEntry{})
		copy(index[ind+1:], index[ind:])
	}

	hash := &nullHash
	if state != StateRemoved {
		var err error
		hash, err = p.createBlob(relPath)
		if err != nil {
			return nil, err
		}
	}

	index[ind] = indexEntry{
		Path:  relPath,
		Hash:  *hash,
		State: state,
		Mtime: mtime,
		Size:  n,
	}
	return index, nil
}

func initIndex(gudPath string) error {
	return dumpIndex(gudPath, []indexEntry{})
}

func loadIndex(gudPath string) ([]indexEntry, error) {
	file, err := os.Open(filepath.Join(gudPath, indexFilePath))

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var index indexFile
	err = gob.NewDecoder(file).Decode(&index)
	if err != nil {
		return nil, err
	}

	if GetVersion() != index.Version { // version does not match
		err := initIndex(gudPath)
		if err != nil {
			return nil, err
		}
		return []indexEntry{}, nil
	}

	return index.Entries, nil
}

func dumpIndex(gudPath string, entries []indexEntry) error {
	file, err := os.Create(filepath.Join(gudPath, indexFilePath))
	if err != nil {
		return err
	}

	err = gob.NewEncoder(file).Encode(indexFile{
		Version: GetVersion(),
		Entries: entries,
	})
	if err != nil {
		return err
	}

	return file.Close()
}

func removeEntry(gudPath string, hash ObjectHash) error {
	if hash != nullHash {
		return os.Remove(objectPath(gudPath, hash))
	}
	return nil
}

func findEntry(entries []indexEntry, relPath string) (int, bool) {
	l := len(entries)
	ind := sort.Search(l, func(i int) bool {
		return relPath <= entries[i].Path
	})

	return ind, ind < l && relPath == entries[ind].Path
}
