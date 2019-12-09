package gud

import (
	"os"
	"path"
	"testing"
)

const testDir string = "test"

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
	info, err := os.Stat(path.Join(testDir, ".gud"))
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
