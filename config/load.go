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

// LoadConfig will parse the os args, read the config file, and return a Config
// struct containing the appropriate values
//
// Config File Location:
// if -config is set:
//		if path is absolute:
//			check if file exists (throw error)
//		else
//			check if file exists in CWD (no error)
//			else if -homedir is set
//					check if file exists in homedir (throw error if it does not)
//					else check if file exists in default factom home (throw error)
// else
//		use default factom home + "factomd.conf"
//		or blank if that does not exist
//
// General order of operation:
// 1. Create default config
// 2. Load config file
// 3. Parse Flags
// 4. Determine "network" from config and flags
// 5. Apply the config, using the settings for appropriate network
// 6. Apply the flags
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
		if fileDoesNotExist(path) {
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
		// TODO yaml "role" support
		//		ext := filepath.Ext(path)
		//		if ext == ".yml" || ext == ".yaml" {
		//			file, err = yamlToIni(path)
		//			if err != nil {
		//				Shutdown(fmt.Errorf("Unable to load yaml file: %v", err))
		//			}
		//		} else {
		file, err = ini.InsensitiveLoad(path)
		if err != nil {
			Shutdown(fmt.Errorf("Unable to load config file: %v", err))
		}
		//		}
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

// determine the network manually so we know which sections to load
// since order isn't guaranteed
func determineNetwork(file *ini.File, flags *Flags) string {
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

// Shutdown is a prettyfied way to stop execution before Factomd is loaded
func Shutdown(err error) {
	fmt.Println(header())
	fmt.Println(WordWrap(fmt.Sprintf("Could not start factomd: %v", err), 80, ""))
	os.Exit(0)
}
