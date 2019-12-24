package gud

import (
	"container/list"
	"crypto/sha1"
	"encoding/gob"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/djherbis/times.v1"
)

const indexFilePath = gudPath + "/index"

type indexEntry struct {
	Name  string
	Hash  [sha1.Size]byte
	Size  int64
	Ctime time.Time
	Mtime time.Time
}

type indexFile struct {
	Version PackageVersion
	Entries []indexEntry
}

func createIndexEntry(rootPath, path string) (*indexEntry, error) {
	relative, err := filepath.Rel(rootPath, path)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(relative, "..") {
		return nil, Error{"Path is not inside the root directory"}
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	spec, err := times.Stat(path)
	if err != nil {
		return nil, err
	}

	hash, err := createBlob(rootPath, relative)
	if err != nil {
		return nil, err
	}

	return &indexEntry{
		Name:  relative,
		Hash:  *hash,
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

	newEntries := make([]indexEntry, 0, len(entries)+files.Len())
	copy(newEntries, entries)

	for e := files.Front(); e != nil; e = e.Next() {
		file := e.Value.(string)

		entry, err := createIndexEntry(rootPath, file)
		if err != nil {
			return err
		}

		ind, found := findEntry(newEntries, entry.Name)
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
		file := e.Value.(string)

		relative, err := filepath.Rel(rootPath, file)
		if err != nil {
			return err
		}

		ind, found := findEntry(entries, relative)
		if found {
			copy(entries[ind:], entries[ind+1:]) // keep the slice sorted
			entries = entries[:len(entries)-1]
		} else {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return Error{string(len(missing)) + " files are not staged"}
	}
	return dumpIndex(rootPath, entries)
}

func loadIndex(rootPath string) ([]indexEntry, error) {
	file, err := os.Open(filepath.Join(rootPath, indexFilePath))
	if err != nil {
		return nil, err
	}

	var index indexFile
	err = gob.NewDecoder(file).Decode(&index)
	if err != nil {
		return nil, err
	}

	err = file.Close()
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

func findEntry(entries []indexEntry, name string) (int, bool) {
	l := len(entries)
	ind := sort.Search(l, func(i int) bool {
		return name <= entries[i].Name
	})

	return ind, ind < l && name == entries[ind].Name
}
