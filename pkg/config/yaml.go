package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func Get(name string) (decode Decode, err error) {
	data, err := os.ReadFile(name)
	if err != nil {
		err = fmt.Errorf("config file error: %v", err)
		return
	}
	fn := func(out interface{}) error {
		return yaml.Unmarshal(data, out)
	}
	return fn, nil
}

type Decode func(out interface{}) error
