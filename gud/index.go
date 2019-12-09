package gud

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/djherbis/times.v1"
)

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

	timespec, err := times.Stat(path)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Name:  relative,
		Size:  info.Size(),
		Mtime: info.ModTime(),
		Ctime: timespec.ChangeTime(),
	}, nil
}

func InitIndex(path string) error {
	return dumpIndex(path, []IndexEntry{})
}

func AddToIndexFile(rootPath, indexPath string, paths []string) error {
	entries, err := loadIndex(indexPath)
	if err != nil {
		return err
	}

	newEntries := make([]IndexEntry, 0, len(entries)+len(paths))
	copy(newEntries, entries)
	for _, path := range paths {
		entry, err := NewIndexEntry(path, rootPath)
		if err != nil {
			return err
		}

		newEntries = append(newEntries, *entry)
	}

	return dumpIndex(indexPath, newEntries)
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
