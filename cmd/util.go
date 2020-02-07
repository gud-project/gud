package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/magsh-2019/2/gud/gud"
)

func getAllFiles() ([]string, error) {

	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Checking if not adding .gud directory
		if !strings.Contains(path, ".gud") && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return files, nil
}

func LoadProject() (*gud.Project, error) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil, err
	}
	p, err := gud.Load(wd)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil, err
	}
	return p, nil
}

func stringToHash(dst *gud.ObjectHash, src string) error{
	_, err := hex.Decode(dst[:], []byte(src))
	if err != nil {
		return err
	}
	return nil
}

const modeMin = "min"
const modeMax = "max"

func checkArgsNum(size , argc int, mode string) error {
	if size > argc && mode != modeMax{
		return fmt.Errorf("not enough arguments in command usage(%d of %d)\n", size, argc)
	}
	if size < argc && mode != modeMin{
		return fmt.Errorf("to many arguments in command usage - %d required\n", argc)
	}
	return nil
}