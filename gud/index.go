package gud

import (
	"encoding/gob"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/djherbis/times.v1"
)

const indexFilePath = gudPath + "/index"

type IndexEntry struct {
	Name  string
	Size  int64
	Ctime time.Time
	Mtime time.Time
}

type indexFile struct {
	Version Version
	Entries []IndexEntry
}

func NewIndexEntry(path, rootPath string) (*IndexEntry, error) {
	relative, err := filepath.Rel(rootPath, path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	spec, err := times.Stat(path)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Name:  relative,
		Size:  info.Size(),
		Mtime: info.ModTime(),
		Ctime: spec.ChangeTime(),
	}, nil
}

func InitIndex(rootPath string) error {
	return dumpIndex(path.Join(rootPath, indexFilePath), []IndexEntry{})
}

func AddToIndex(rootPath string, paths []string) error {
	indexPath := path.Join(rootPath, indexFilePath)
	entries, err := loadIndex(indexPath)
	if err != nil {
		return err
	}

	newEntries := make([]IndexEntry, 0, len(entries)+len(paths))
	copy(newEntries, entries)

	for _, file := range paths {
		// TODO: if file is a directory, create new entries recursively
		entry, err := NewIndexEntry(file, rootPath)
		if err != nil {
			return err
		}

		ind := findEntry(newEntries, entry.Name)
		if ind >= len(newEntries) || file != newEntries[ind].Name { // file is not yet added
			newEntries = append(newEntries, IndexEntry{})
			copy(newEntries[ind+1:], newEntries[ind:]) // keep the slice sorted
		}
		newEntries[ind] = *entry // update entry if the file was already added
	}

	return dumpIndex(indexPath, newEntries)
}

func RemoveFromIndex(rootPath string, paths []string) error {
	indexPath := path.Join(rootPath, indexFilePath)
	entries, err := loadIndex(indexPath)
	if err != nil {
		return err
	}

	missing := make([]string, 0, len(paths))

	for _, file := range paths {
		relative, err := filepath.Rel(rootPath, file)
		if err != nil {
			return err
		}

		ind := findEntry(entries, relative)
		if ind < len(entries) && relative == entries[ind].Name { // file is found
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

func findEntry(entries []IndexEntry, name string) int {
	return sort.Search(len(entries), func(i int) bool {
		return name == entries[i].Name
	})
}
