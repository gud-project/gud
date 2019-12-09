package gud

import (
	"encoding/gob"
	"os"
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

func NewIndexEntry(path string) (*IndexEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &IndexEntry{
		Name:  path,
		Size:  info.Size(),
		Mtime: info.ModTime(),
		Ctime: times.Get(info).ChangeTime(),
	}, nil
}

func InitIndex(path string) error {
	return dumpIndex(path, []IndexEntry{})
}

func AddToIndexFile(indexPath string, paths []string) error {
	entries, err := loadIndex(indexPath)
	if err != nil {
		return err
	}

	newEntries := make([]IndexEntry, 0, len(entries)+len(paths))
	copy(newEntries, entries)
	for i, path := range paths {
		entry, err := NewIndexEntry(path)
		if err != nil {
			return err
		}

		newEntries[len(entries)+i] = *entry
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
