package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/FactomProject/factomd/util"
	"github.com/go-ini/ini"
)

var relativeRegExp = regexp.MustCompile(`^([A-Za-z]:)|(~?(\.\.?)?[/\\])`)

func fileDoesNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func LoadConfig() Config {
	flags, err := ParseOSFlags()
	if err != nil {
		Shutdown(err)
	}

	_, ok := flags.GetS("help", "h", "?")
	if ok {
		fmt.Println(GetUsage())
		os.Exit(0)
	}

	path, ok := flags.GetS("config", "c")
	home, hok := flags.GetS("homedir", "h")

	// TODO use path.FilePath instead of this

	if ok { // a config file path was passed
		if relativeRegExp.MatchString(path) { // is relative, convert to absolute
			if fileDoesNotExist(path) { // file does not exist in CWD
				if hok {
					path = home + path // they gave us a new homedir
				} else {
					path = util.GetHomeDir() + "/.factom/m2/" + path // use default homedir
				}
			}
		}

		// at this point, path is absolute
		if _, err := os.Stat(path); os.IsNotExist(err) {
			Shutdown(fmt.Errorf("could not read file: %v", err))
		}
	} else {
		path = util.GetConfigFilename("m2")
	}

	//ini.Load()
	return Config{}
}

func parseConfig(file *ini.File, flags *Flags) (Config, error) {
	c := DefaultConfig()

	network := file.Section("Factomd").Key("Network").String()

	if !fNetwork(network) {
		return c, fmt.Errorf("Network name \"%s\" could not be parsed. Use alphanumeric characters and _ only", network)
	}

	walk(&c, func(category reflect.StructField, field reflect.StructField, val reflect.Value) error {

		return nil
	})
	return c, nil
}

func Shutdown(err error) {
	fmt.Println("Could not start factomd:", err)
	os.Exit(1)
}
