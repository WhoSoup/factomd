package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/FactomProject/factomd/util"
	"github.com/go-ini/ini"
)

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

	// turn homedir into an absolute path
	if hok && !filepath.IsAbs(home) {
		home, err = filepath.Abs(home)
		if err != nil {
			Shutdown(fmt.Errorf("Unable to parse the path of -homedir: %v", err))
		}
	}

	// TODO use path.FilePath instead of this
	if ok { // a config file path was passed
		if !filepath.IsAbs(path) { // is relative, convert to absolute
			if fileDoesNotExist(path) { // file does not exist in CWD
				if hok { // custom homedir
					path = filepath.Join(home, path)
				} else { // default homedir
					path = filepath.Join(util.GetHomeDir(), ".factom/m2", path)
				}
			} else {
				path, err = filepath.Abs(path)
				if err != nil {
					Shutdown(fmt.Errorf("Unable to parse the -config path: %v", err))
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

	file, err := ini.InsensitiveLoad(path)
	if err != nil {
		Shutdown(fmt.Errorf("Unable to load config file: %v", err))
	}

	cfg, err := parseConfig(file, flags)

	if err != nil {
		Shutdown(fmt.Errorf("Problem reading configuration\n%v", err))
	}

	if unused := flags.Unused(); len(unused) > 0 {
		Shutdown(fmt.Errorf("Unknown command line parameter: %s", unused[0]))
	}

	return cfg
}

func parseConfig(file *ini.File, flags *Flags) (Config, error) {
	c := DefaultConfig()

	// determine the network manually so we know which sections to load
	// since order isn't guaranteed
	network, ok := flags.GetS("network", "n")
	if !ok {
		if file.Section("Factomd").HasKey("network") {
			network = file.Section("Factomd").Key("Network").String()
		} else {
			network = c.Factomd.Network
		}
	}

	if err := stringFTag("network", network); err != nil {
		return c, err
	}
	c.Factomd.Network = network

	// order of priority from least to highest:
	// default, config global, config network specific, command line long, command line short
	err := walk(&c, func(category reflect.StructField, field reflect.StructField, val reflect.Value) error {
		// there's a short tag
		if short, ok := field.Tag.Lookup("short"); ok {
			if short, ok = flags.Get(short); ok {
				err := set(val, short, field.Tag)
				if err != nil {
					return fmt.Errorf("invalid input for command line option \"-%s\":\n%v", short, err)
				}
				return nil
			}
		}

		// normal command line
		if long, ok := flags.Get(field.Name); ok {
			err := set(val, long, field.Tag)
			if err != nil {
				return fmt.Errorf("invalid input for command line option \"-%s\":\n%v", field.Name, err)
			}
			return nil
		}

		section := fmt.Sprintf("%s.%s", category.Name, network) // ini package automatically handles inheritance
		if file.Section(section).HasKey(field.Name) {
			err := set(val, file.Section(section).Key(field.Name).String(), field.Tag)
			if err != nil {
				return fmt.Errorf("invalid value for \"%s.%s\" in the config file:\n%v", section, field.Name, err)
			}
		}

		// default is already set
		return nil
	})

	return c, err
}

func Shutdown(err error) {
	fmt.Println(header())
	fmt.Println(WordWrap(fmt.Sprintf("Could not start factomd: %v", err), 80, ""))
	os.Exit(0)
}
