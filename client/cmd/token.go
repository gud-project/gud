package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

const tokenPath = "path"

func saveToken(name, token string) error{
	f, err := os.Create(filepath.Join(tokenPath, name))
	if err != nil {
		return fmt.Errorf("Failed to create token file: %s\n", err.Error())
	}
	defer f.Close()

	_, err = f.Write([]byte(token))
	if err != nil {
		return fmt.Errorf("Failed to write token: %s\n", err.Error())
	}
	return nil
}

func loadToken(name string) (string, error) {
	return "yay", nil
}
