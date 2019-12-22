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

type IndexEntry struct {
	Name  string
	Hash  [sha1.Size]byte
	Size  int64
	Ctime time.Time
	Mtime time.Time
}

type indexFile struct {
	Version PackageVersion
	Entries []IndexEntry
}

func CreateIndexEntry(rootPath, path string) (*IndexEntry, error) {
	relative, err := filepath.Rel(rootPath, path)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(relative, "..") {
		return nil, Error{"path is not inside the root directory"}
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	spec, err := times.Stat(path)
	if err != nil {
		return nil, err
	}

	hash, err := CreateBlob(rootPath, relative)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Name:  relative,
		Hash:  *hash,
		Size:  info.Size(),
		Mtime: info.ModTime(),
		Ctime: spec.ChangeTime(),
	}, nil
}

func InitIndex(rootPath string) error {
	return dumpIndex(filepath.Join(rootPath, indexFilePath), []IndexEntry{})
}

func AddToIndex(rootPath string, paths []string) error {
	// TODO: handle renames
	indexPath := filepath.Join(rootPath, indexFilePath)
	entries, err := loadIndex(indexPath)
	if err != nil {
		return err
	}

	files, err := walkFiles(paths)
	if err != nil {
		return err
	}

	newEntries := make([]IndexEntry, 0, len(entries)+files.Len())
	copy(newEntries, entries)

	for e := files.Front(); e != nil; e = e.Next() {
		file := e.Value.(string)

		entry, err := CreateIndexEntry(rootPath, file)
		if err != nil {
			return err
		}

		ind, found := findEntry(newEntries, entry.Name)
		if !found { // file is not yet added
			newEntries = append(newEntries, IndexEntry{})
			copy(newEntries[ind+1:], newEntries[ind:]) // keep the slice sorted
		}
		newEntries[ind] = *entry // update entry if the file was already added
	}

	return dumpIndex(indexPath, newEntries)
}

func RemoveFromIndex(rootPath string, paths []string) error {
	indexPath := filepath.Join(rootPath, indexFilePath)
	entries, err := loadIndex(indexPath)
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
	return dumpIndex(indexPath, entries)
}

func loadIndex(path string) ([]IndexEntry, error) {
	file, err := os.Open(path)
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

	if GetVersion() != index.Version {
		panic("Outdated gud version") // TODO: clean up index file
	}

	return index.Entries, nil
}

func dumpIndex(path string, entries []IndexEntry) error {
	file, err := os.Create(path)
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

func findEntry(entries []IndexEntry, name string) (int, bool) {
	l := len(entries)
	ind := sort.Search(l, func(i int) bool {
		return name <= entries[i].Name
	})

	return ind, ind < l && name == entries[ind].Name
}
