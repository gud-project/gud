package gud

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testFile string = "testFile"

func TestInitObjectsDir(t *testing.T) {
	defer clearTest()

	_ = os.Mkdir(filepath.Join(testDir, gudPath), os.ModeDir)
	_, err := InitObjectsDir(testDir)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(filepath.Join(testDir, objectsDirPath)); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCreateBlob(t *testing.T) {
	defer clearTest()

	_, _ = Start(testDir)
	_ = ioutil.WriteFile(filepath.Join(testDir, testFile), []byte("hello\nthis is a test"), 0644)

	hash, err := CreateBlob(testDir, testFile)
	if err != nil {
		t.Error(err)
	}

	if _, err = os.Stat(filepath.Join(testDir, objectsDirPath, hex.EncodeToString(hash[:]))); os.IsNotExist(err) {
		t.Error(err)
	}
}
