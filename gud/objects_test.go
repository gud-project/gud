package gud

import (
	"os"
	"path"
	"testing"
)

const TestFile string = "testFile"

func TestInitObjectsDir(t *testing.T) {
	err := InitObjectsDir(testDir)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(path.Join(testDir, ObjectsDirName)); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCrerteBlob(t *testing.T) {
	f, err := os.Create(TestFile)
	if err != nil {
		t.Error(err)
	}
	_, err = f.Write([]byte("hello\nthis is a test"))
	if err != nil {
		t.Error(err)
	}

	_, err = CreateBlob(TestFile)
	if err != nil {
		t.Error(err)
	}

	err = f.Close()
	if err != nil {
		t.Error(err)
	}
}
