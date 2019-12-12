package gud

import (
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

	testPath := path.Join(testDir, testFile)

	f, err := os.Create(testPath)
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello\nthis is a test"))
	if err != nil {
		t.Error(err)
	}

	_, err = CreateBlob(testPath)
	if err != nil {
		t.Error(err)
	}

	err = f.Close()
	if err != nil {
		t.Error(err)
	}
}
