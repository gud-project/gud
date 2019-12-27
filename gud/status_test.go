package gud

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestProject_Status(t *testing.T) {
	defer clearTest()

	const testFile = "foo.txt"
	testPath := filepath.Join(testDir, testFile)
	data := []byte("random test data")

	p, _ := Start(testDir)
	_ = ioutil.WriteFile(testPath, data, 0644)

	err := p.Status(
		func(relPath string, state FileState) error {
			return fmt.Errorf("file %s (%d) is not tracked", relPath, state)
		},
		func(relPath string, state FileState) error {
			if relPath != testFile || state != StateNew {
				return fmt.Errorf("unexpected %s (%d)", relPath, state)
			}
			return nil
		},
	)
	if err != nil {
		t.Error(err)
	}

	_ = addToIndex(testDir, []string{testPath})

	err = p.Status(
		func(relPath string, state FileState) error {
			if relPath != testFile || state != StateNew {
				return fmt.Errorf("unexpected %s (%d)", relPath, state)
			}
			return nil
		},
		func(relPath string, state FileState) error {
			return fmt.Errorf("file %s (%d) is tracked", relPath, state)
		},
	)
	if err != nil {
		t.Error(err)
	}

	_, _ = p.Save("added test file")

	err = p.Status(
		func(relPath string, state FileState) error {
			return fmt.Errorf("unexpected tracked %s (%d)", relPath, state)
		},
		func(relPath string, state FileState) error {
			return fmt.Errorf("unexpected untracked %s (%d)", relPath, state)
		},
	)
	if err != nil {
		t.Error(err)
	}
}
