package gud

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"path/filepath"
	)

const configFile = ".gud/config.toml"

type Config struct {
	Name string
	ProjectName string
	Token string
	ServerDomain string
	Checkpoints int
	AutoPush bool
}

func (p *Project)ConfigInit() error{
	config := Config{"", filepath.Base(p.Path), "", "localhost", 3, true}
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(p.Path, configFile))
	if err != nil {
		return fmt.Errorf("Failed to create configuration file: %s\n", err.Error())
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("Failed to write configuration: %s\n", err.Error())
	}
	return nil
}

func (p *Project)WriteConfig(config Config) error {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	f, err :=  os.OpenFile(filepath.Join(p.Path, configFile), os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (p *Project)LoadConfig(config *Config) error {
	b ,err := p.ReadConfig()
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, config)
}

func (p *Project)ReadConfig() ([]byte, error) {
	f, err := os.Open(filepath.Join(p.Path, configFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}