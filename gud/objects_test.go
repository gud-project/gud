package gud

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestInitObjectsDir(t *testing.T) {
	defer clearTest()

	_ = os.Mkdir(filepath.Join(testDir, gudPath), dirPerm)
	err := initObjectsDir(testDir)
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

	hash, err := createBlob(testDir, testFile)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = os.Stat(filepath.Join(testDir, objectsDirPath, hex.EncodeToString(hash[:]))); os.IsNotExist(err) {
		t.Error(err)
	}
}
