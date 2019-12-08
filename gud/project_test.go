package gud

import (
	"os"
	"path"
	"testing"
)

const testDir string = "test"

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
	// Checks if project already exist
	_, err := os.Stat(path.Join(testDir, ".gud"))
	if os.IsNotExist(err) {
		// Create gud project to load
		_, err := Start(testDir)
		if err != nil {
			t.Error(err)
		}
	}

	// Runs function Load
	_, err = Load(testDir)
	if err != nil {
		t.Error(err)
	}
}
