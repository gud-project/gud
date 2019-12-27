package gud

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestAddToIndex(t *testing.T) {
	defer clearTest()

	const testFile = "foo.txt"
	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	_, _ = Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)

	err := addToIndex(testDir, []string{testPath})
	if err != nil {
		t.Error(err)
	}

	entries, err := loadIndex(testDir)
	if err != nil {
		t.Error(err)
	}

	if len(entries) != 1 || entries[0].Name != testFile || entries[0].Size != int64(len(data)) {
		t.Fail()
	}
}

func TestRemoveFromIndex(t *testing.T) {
	defer clearTest()

	const testFile = "foo.txt"
	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	_, _ = Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)
	_ = addToIndex(testDir, []string{testPath})

	err := removeFromIndex(testDir, []string{testPath})
	if err != nil {
		t.Error(err)
	}

	entries, err := loadIndex(testDir)
	if err != nil {
		t.Error(err)
	}

	if len(entries) > 0 {
		t.Fail()
	}
}

func TestRemoveFromProject(t *testing.T) {
	defer clearTest()

	const testFile = "foo.txt"
	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	p, _ := Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)
	_ = addToIndex(testDir, []string{testPath})

	err := removeFromProject(testDir, []string{testPath})
	if err != nil {
		t.Error(err)
	}

	entries, err := loadIndex(testDir)
	if err != nil {
		t.Error(err)
	}
	if len(entries) > 0 {
		t.Error("Index entry was not removed")
	}

	_ = p.Add(testDir, testPath)
	_, _ = p.Save("add test file")

	err = removeFromProject(testDir, []string{testPath})
	if err != nil {
		t.Error(err)
	}

	entries, err = loadIndex(testDir)
	if err != nil {
		t.Error(err)
	}

	if len(entries) != 1 || entries[0].Name != testFile || entries[0].State != StateRemoved {
		t.Error("Index entry was not added")
	}
}
