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

	err := AddToIndex(testDir, []string{testPath})
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
	_ = AddToIndex(testDir, []string{testPath})

	err := RemoveFromIndex(testDir, []string{testPath})
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
