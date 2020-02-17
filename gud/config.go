package gud

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const configFile = "config.toml"

type Config struct {
	Name         string
	ProjectName  string
	Token        string
	ServerDomain string
	Checkpoints  int
	AutoPush     bool
}

func (p *Project) ConfigInit() (err error) {
	config := Config{"", filepath.Base(p.Path), "", "localhost", 3, false}
	b, err := toml.Marshal(config)
	if err != nil {
		return
	}

	f, err := os.Create(filepath.Join(p.gudPath, configFile))
	if err != nil {
		return fmt.Errorf("Failed to create configuration file: %s\n", err.Error())
	}
	defer func() {
		cerr := f.Close()
		if err != nil {
			err = cerr
		}
	}()

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("Failed to write configuration: %s\n", err.Error())
	}
	return
}

func (p *Project) WriteConfig(config Config) (err error) {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	f, err :=  os.Create(filepath.Join(p.gudPath, configFile))
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if err != nil {
			err = cerr
		}
	}()

	_, err = f.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) LoadConfig(config *Config) error {
	b ,err := p.ReadConfig()
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, config)
}

func (p *Project) ReadConfig() ([]byte, error) {
	f, err := os.Open(filepath.Join(p.gudPath, configFile))
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
