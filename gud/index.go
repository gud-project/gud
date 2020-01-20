package gud

import (
	"container/list"
	"encoding/gob"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/djherbis/times.v1"
)

const indexFilePath = gudPath + "/index"

type FileState int

const (
	StateNew FileState = iota
	StateRemoved
	StateModified
	StateConflict
)

type indexEntry struct {
	Name  string
	Hash  objectHash
	State FileState
	Size  int64
	Ctime time.Time
	Mtime time.Time
}

type indexFile struct {
	Version PackageVersion
	Entries []indexEntry
}

func createIndexEntry(rootPath, relPath string) (*indexEntry, error) {
	if strings.HasPrefix(relPath, "..") {
		return nil, Error{rootPath + " is not inside the root directory"}
	}

	prevHash, err := findObject(rootPath, relPath)
	if err != nil {
		return nil, err
	}

	var state FileState
	if prevHash != nil {
		unchanged, err := compareToObject(rootPath, relPath, *prevHash)
		if err != nil {
			return nil, err
		}
		if unchanged {
			return nil, AddedUnmodifiedFileError
		}
		state = StateModified
	} else {
		state = StateNew
	}

	fullPath := filepath.Join(rootPath, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	spec, err := times.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	hash, err := createBlob(rootPath, relPath)
	if err != nil {
		return nil, err
	}

	return &indexEntry{
		Name:  relPath,
		Hash:  *hash,
		State: state,
		Size:  info.Size(),
		Mtime: info.ModTime(),
		Ctime: spec.ChangeTime(),
	}, nil
}

func initIndex(rootPath string) error {
	return dumpIndex(rootPath, []indexEntry{})
}

func addToIndex(rootPath string, paths []string) error {
	// TODO: handle renames
	entries, err := loadIndex(rootPath)
	if err != nil {
		return err
	}

	files, err := walkFiles(paths)
	if err != nil {
		return err
	}

	newEntries := make([]indexEntry, len(entries), len(entries)+files.Len())
	copy(newEntries, entries)

	for e := files.Front(); e != nil; e = e.Next() {
		fullPath := e.Value.(string)

		relPath, err := filepath.Rel(rootPath, fullPath)
		if err != nil {
			return err
		}

		ind, found := findEntry(newEntries, relPath)
		entry, err := createIndexEntry(rootPath, relPath)
		if err != nil {
			if err == AddedUnmodifiedFileError && found && newEntries[ind].State == StateRemoved { // remode then add
				copy(newEntries[ind:], newEntries[ind+1:])
				newEntries = newEntries[:len(newEntries)-1]
				continue
			}
			return err
		}

		if !found { // file is not yet added
			newEntries = append(newEntries, indexEntry{})
			copy(newEntries[ind+1:], newEntries[ind:]) // keep the slice sorted
		}
		newEntries[ind] = *entry // update entry if the file was already added
	}

	return dumpIndex(rootPath, newEntries)
}

func removeFromIndex(rootPath string, paths []string) error {
	entries, err := loadIndex(rootPath)
	if err != nil {
		return err
	}

	files, err := walkFiles(paths)
	if err != nil {
		return nil
	}

	missing := make([]string, 0, files.Len())

	for e := files.Front(); e != nil; e = e.Next() {
		path := e.Value.(string)

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		ind, found := findEntry(entries, relPath)
		if found {
			copy(entries[ind:], entries[ind+1:]) // keep the slice sorted
			entries = entries[:len(entries)-1]
		} else {
			missing = append(missing, path)
		}
	}

	if len(missing) > 0 {
		return Error{string(len(missing)) + " files are not staged"}
	}
	return dumpIndex(rootPath, entries)
}

func removeFromProject(rootPath string, paths []string) error {
	entries, err := loadIndex(rootPath)
	if err != nil {
		return err
	}

	files, err := walkFiles(paths) // TODO: walk objects instead of files?
	if err != nil {
		return nil
	}

	missing := make([]string, 0, files.Len())
	newEntries := make([]indexEntry, len(entries), len(entries)+files.Len())
	copy(newEntries, entries)

	for e := files.Front(); e != nil; e = e.Next() {
		path := e.Value.(string)

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		ind, found := findEntry(newEntries, relPath)
		if found {
			entry := &newEntries[ind]
			if entry.State == StateNew {
				copy(newEntries[ind:], newEntries[ind+1:])
				newEntries = newEntries[:len(newEntries)-1]
			} else {
				entry.State = StateRemoved
				entry.Hash = nullHash
			}

		} else {
			// check that the file existed before removing it
			prevHash, err := findObject(rootPath, relPath)
			if err != nil {
				return err
			}
			if prevHash == nil {
				missing = append(missing, path)
			} else {
				newEntries = append(newEntries, indexEntry{})
				copy(newEntries[ind+1:], newEntries[ind:]) // keep the slice sorted
				newEntries[ind] = indexEntry{
					Name:  relPath,
					Hash:  nullHash,
					State: StateRemoved,
					Ctime: time.Now(),
					Mtime: time.Now(),
				}
			}
		}
	}

	if len(missing) > 0 {
		return Error{string(len(missing)) + " files are not tracked"}
	}
	return dumpIndex(rootPath, newEntries)
}

func loadIndex(rootPath string) ([]indexEntry, error) {
	file, err := os.Open(filepath.Join(rootPath, indexFilePath))

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
		err := initIndex(rootPath)
		if err != nil {
			return nil, err
		}
		return []indexEntry{}, nil
	}

	return index.Entries, nil
}

func dumpIndex(rootPath string, entries []indexEntry) error {
	file, err := os.Create(filepath.Join(rootPath, indexFilePath))
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

func walkFiles(paths []string) (*list.List, error) {
	files := list.New()

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files.PushFront(path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func findEntry(entries []indexEntry, relPath string) (int, bool) {
	l := len(entries)
	ind := sort.Search(l, func(i int) bool {
		return relPath <= entries[i].Name
	})

	return ind, ind < l && relPath == entries[ind].Name
}
