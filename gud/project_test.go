package gud

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const testDir string = "test"
const testFile string = "testFile"

func clearTest() {
	err := os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(testDir, os.ModeDir)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	// Creates test directory
	err := os.Mkdir(testDir, os.ModeDir)
	if err != nil {
		panic(err)
	}

	// Runs the functions
	rt := m.Run()

	// Deletes test directory
	err = os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}

	os.Exit(rt)
}

func TestStart(t *testing.T) {
	defer clearTest()

	// Runs function Start
	_, err := Start(testDir)
	if err != nil {
		t.Error(err)
	}

	// Checks if dir created
	info, err := os.Stat(filepath.Join(testDir, ".gud"))
	if os.IsNotExist(err) || !info.IsDir() {
		t.Error(err)
	}
}

func TestLoad(t *testing.T) {
	defer clearTest()

	_, _ = Start(testDir)

	_, err := Load(testDir)
	if err != nil {
		t.Error(err)
	}
}

func TestProject_Save(t *testing.T) {
	defer clearTest()

	testPath := filepath.Join(testDir, testFile)
	p, _ := Start(testDir)
	_ = ioutil.WriteFile(testPath, []byte("hello\nthis is a test"), 0644)

	_ = p.Add(testPath)
	version, err := p.Save("add testFile")
	if err != nil {
		t.Error(err)
	}

	var tree Tree
	err = LoadTree(testDir, version.Tree, &tree)
	if err != nil {
		t.Error(err)
	}

	if len(tree) != 1 {
		t.FailNow()
	}

	obj := tree[0]
	if obj.Name != testFile || obj.Type != typeBlob {
		t.FailNow()
	}
}
