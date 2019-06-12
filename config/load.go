package config

import (
	"fmt"
	"reflect"

	"github.com/go-ini/ini"
)

func LoadConfig() Config {
	return Config{}
}

func parseConfig(file *ini.File) (Config, error) {
	c := DefaultConfig()

	network := file.Section("Factomd").Key("Network").String()

	if !fNetwork(network) {
		return c, fmt.Errorf("Network name \"%s\" could not be parsed. Use alphanumeric characters and _ only", network)
	}

	apply(&c, func(category reflect.StructField, field reflect.StructField, val reflect.Value) error {

		return nil
	})
	return c, nil
}
