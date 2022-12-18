package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func ProvideConfig() (*OwlBotConfig, error) {
	yfile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}

	data := OwlBotConfig{}
	err = yaml.Unmarshal(yfile, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}