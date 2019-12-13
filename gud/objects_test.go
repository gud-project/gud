package gud

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const testFile string = "testFile"

func TestInitObjectsDir(t *testing.T) {
	defer clearTest()

	_ = os.Mkdir(path.Join(testDir, gudPath), os.ModeDir)
	err := InitObjectsDir(testDir)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(path.Join(testDir, objectsDirPath)); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCreateBlob(t *testing.T) {
	defer clearTest()

	_, _ = Start(testDir)
	_ = ioutil.WriteFile(path.Join(testDir, testFile), []byte("hello\nthis is a test"), 0644)

	hash, err := CreateBlob(testDir, testFile)
	if err != nil {
		t.Error(err)
	}

	if _, err = os.Stat(path.Join(testDir, objectsDirPath, string(hash[:]))); os.IsNotExist(err) {
		t.Error(err)
	}
}
