package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestInitObjectsDir(t *testing.T) {
	defer clearTest()

	gudPath := filepath.Join(testDir, DefaultPath)
	_ = os.Mkdir(gudPath, dirPerm)
	err := initObjectsDir(gudPath)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(filepath.Join(gudPath, objectsPath)); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestProject_createBlob(t *testing.T) {
	defer clearTest()

	p, _ := Start(testDir)
	_ = ioutil.WriteFile(filepath.Join(testDir, testFile), []byte("hello\nthis is a test"), 0644)

	hash, err := p.createBlob(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = os.Stat(objectPath(p.gudPath, *hash)); os.IsNotExist(err) {
		t.Error(err)
	}
}
