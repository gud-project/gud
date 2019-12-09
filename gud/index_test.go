package gud

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestAddToIndexFile(t *testing.T) {
	defer clearTest()

	const testFile = "foo.txt"
	testPath := path.Join(testDir, testFile)
	data := []byte("random test data")

	_, _ = Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)

	err := AddToIndexFile(testDir, []string{testPath})
	if err != nil {
		t.Error(err)
	}

	entries, err := loadIndex(path.Join(testDir, ".gud/index"))
	if err != nil {
		t.Error(err)
	}

	if len(entries) != 1 || entries[0].Name != testFile || entries[0].Size != int64(len(data)) {
		t.Fail()
	}
}
