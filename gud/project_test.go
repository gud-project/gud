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

	err = os.Mkdir(testDir, dirPerm)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	_ = os.RemoveAll(testDir)
	// Creates test directory
	err := os.Mkdir(testDir, dirPerm)
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

	tree, err := loadTree(p.gudPath, version.TreeHash)
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
	_, beforeFirst, err := p.Prev(*firstVersion)
	if beforeFirst != nil || err == nil {
		t.Fail()
	}

	_ = ioutil.WriteFile(testPath, []byte("hello\nthis is a test"), 0644)

	_ = p.Add(testPath)
	secondVersion, _ := p.Save("add testFile")

	_, prev, err := p.Prev(*secondVersion)
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
	if firstBranch != FirstBranchName {
		t.Error("first branch name incorrect")
	}
}

func TestProject_Checkpoint(t *testing.T) {
	defer clearTest()

	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	p, _ := Start(testDir)
	err := p.Checkpoint("gud start")
	if err != nil {
		t.Fatal("failed checkpoint after start:", err)
	}

	_ = ioutil.WriteFile(testPath, data, 0644)
	_ = p.Add(testPath)
	err = p.Checkpoint("gud add")
	if err != nil {
		t.Fatal("failed checkpoint after add:", err)
	}

	_, err = p.Save("add file")
	err = p.Undo()
	if err != nil {
		t.Fatal("failed to undo save:", err)
	}

	version, err := p.CurrentVersion()
	if err != nil {
		t.Fatal("failed to load current version:", err)
	}
	if version.Message != initialCommitName {
		t.Fatal("current version not undo'ed")
	}

	err = p.Undo()
	if err != nil {
		t.Fatal("failed to undo add:", err)
	}

	index, err := loadIndex(p.gudPath)
	if err != nil {
		t.Fatal("failed to load index:", err)
	}

	if len(index) > 0 {
		t.Fatal("action not undo'ed properly")
	}
}
