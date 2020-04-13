package gud

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const localConfigPath = "config.toml"
const defaultDomainServer = "https://gud.codes"


type Config struct {
	ProjectName string
	OwnerName string
	Checkpoints  int
	AutoPush     bool
}

type GlobalConfig struct {
	Name, Token, ServerDomain string
}

func (config GlobalConfig) GetPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal( err )
	}
	return filepath.Join(usr.HomeDir, ".gudConfig.toml")
}

func (p *Project) ConfigInit() (err error) {
	config := Config{filepath.Base(p.Path), "", 3, false}
	b, err := toml.Marshal(config)
	if err != nil {
		return
	}

	f, err := os.Create(filepath.Join(p.gudPath, localConfigPath))
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
	return WriteConfig(config, filepath.Join(p.gudPath, localConfigPath))
}

func WriteConfig(config interface{}, path string) (err error) {
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	f, err :=  os.Create(path)
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

func LoadConfig(config interface{}, path string) error {
	b ,err := ReadConfig(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(b, config)
}


func (p *Project) ReadConfig() ([]byte, error) {
	return ReadConfig(filepath.Join(p.gudPath, localConfigPath))
}

func ReadConfig(path string) ([]byte, error) {
	_, err := os.Stat(path)
	var f *os.File
	if os.IsNotExist(err) {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}

		err = WriteConfig(GlobalConfig{"", "", defaultDomainServer}, GlobalConfig{}.GetPath())
		if err != nil {
			return nil, err
		}

	} else {
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return b, nil
}