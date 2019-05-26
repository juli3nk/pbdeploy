package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func New(filename string) (*Config, error) {
	if _, err := os.Lstat(filename); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := new(Config)

	if err = yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if len(config.Repository.Provider) == 0 {
		return nil, fmt.Errorf("Repository Provider not set")
	}
	if len(config.Repository.Username) == 0 {
		return nil, fmt.Errorf("Repository Username not set")
	}
	for i, pack := range config.Packages {
		if len(pack.Type) == 0 {
			return nil, fmt.Errorf("Package Type not set")
		}
		if len(pack.Name) == 0 {
			return nil, fmt.Errorf("Package Name not set")
		}
		if len(pack.Org) == 0 {
			config.Packages[i].Org = config.Repository.Username
		}
		if len(pack.SrcPath) == 0 {
			return nil, fmt.Errorf("Package SrcPath not set")
		}
	}

	return config, nil
}
