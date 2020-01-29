package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

const tokenName = ".gudToken"

func getTokenPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal( err )
	}
	return filepath.Join(usr.HomeDir, tokenName)
}

func saveToken(token string) error{
	f, err := os.Create(getTokenPath())
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

func loadToken() (string, error) {
	f, err := os.Open(getTokenPath())
	if err != nil {
		return "", errors.New("Failed to load token\n")
	}
	defer f.Close()
	return "yay", nil
}

func deleteToken() {
	err := os.Remove(getTokenPath())
	if err != nil {
		print(err.Error())
	}
}