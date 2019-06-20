package config

import (
	"fmt"
	"os"
	"path/filepath"

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
		path = util.GetConfigFilename("m2") // the default config is optional
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = ""
		}
	}

	// order of priority from least to highest:
	// default, config global, config network specific, command line long, command line short

	cfg := DefaultConfig()

	// load file
	var file *ini.File
	if path != "" {
		file, err = ini.InsensitiveLoad(path)
		if err != nil {
			Shutdown(fmt.Errorf("Unable to load config file: %v", err))
		}
	} else {
		file = ini.Empty()
	}

	// determine network from file and flags

	network := determineNetwork(file, flags)
	if network == "" {
		network = cfg.Factomd.Network
	}

	// CONFIG
	err = cfg.addConfig(file, network)
	if err != nil {
		Shutdown(fmt.Errorf("Problem reading configuration\n%v", err))
	}

	// FLAGS
	err = cfg.addFlags(flags)
	if err != nil {
		Shutdown(fmt.Errorf("Problem reading command line flags\n%v", err))
	}

	if unused := flags.Unused(); len(unused) > 0 {
		Shutdown(fmt.Errorf("Unknown command line parameter: %s", unused[0]))
	}

	return *cfg
}

func determineNetwork(file *ini.File, flags *Flags) string {
	// determine the network manually so we know which sections to load
	// since order isn't guaranteed
	network, ok := flags.GetS("network", "n")
	if ok {
		return network
	}

	if file.Section("Factomd").HasKey("network") {
		// note: the content of this setting will be verified during parseConfig()
		return file.Section("Factomd").Key("Network").String()
	}

	return ""
}

func Shutdown(err error) {
	fmt.Println(header())
	fmt.Println(WordWrap(fmt.Sprintf("Could not start factomd: %v", err), 80, ""))
	os.Exit(0)
}
