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

	current, err := p.CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if current.TreeHash != version.TreeHash {
		t.Error("CurrentVersion() did not return the latest version")
	}

	tree, err := loadTree(testDir, version.TreeHash)
	if err != nil {
		t.Fatal(err)
	}

	if len(tree) != 1 {
		t.FailNow()
	}

	obj := tree[0]
	if obj.Name != testFile || obj.Type != typeBlob {
		t.FailNow()
	}
}

func TestProject_Prev(t *testing.T) {
	defer clearTest()

	testPath := filepath.Join(testDir, testFile)
	p, _ := Start(testDir)
	firstVersion, _ := p.CurrentVersion()
	beforeFirst, err := p.Prev(*firstVersion)
	if beforeFirst != nil || err == nil {
		t.Fail()
	}

	_ = ioutil.WriteFile(testPath, []byte("hello\nthis is a test"), 0644)

	_ = p.Add(testPath)
	secondVersion, _ := p.Save("add testFile")

	prev, err := p.Prev(*secondVersion)
	if err != nil {
		t.Fatal(err)
	}

	if prev.TreeHash != firstVersion.TreeHash {
		t.FailNow()
	}
}

func TestProject_CurrentBranch(t *testing.T) {
	defer clearTest()

	p, _ := Start(testDir)

	firstBranch, err := p.CurrentBranch()
	if err != nil {
		t.Error(err)
	}
	if firstBranch != firstBranchName {
		t.Error("first branch name incorrect")
	}
}
