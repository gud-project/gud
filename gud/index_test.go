package gud

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestProject_Add(t *testing.T) {
	defer clearTest()

	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	p, _ := Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)

	err := p.Add(testPath)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := p.getIndex()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 || entries[0].Path != testFile {
		t.Fail()
	}
}

func TestProject_Remove(t *testing.T) {
	defer clearTest()

	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	p, _ := Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)
	_ = p.Add(testPath)

	err := p.Remove(testPath)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := p.getIndex()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Error("Index entry was not removed")
	}

	_ = p.Add(testPath)
	_, _ = p.Save("add test file")

	err = p.Remove(testPath)
	if err != nil {
		t.Fatal(err)
	}

	entries, err = p.getIndex()
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 || entries[0].Path != testFile || entries[0].State != StateRemoved {
		t.Error("Index entry was not added")
	}
}
