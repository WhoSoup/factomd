package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Config map[string]ConfigGroup

type ConfigGroup map[string]Setting

type Setting struct {
	Short       string // optional short identifier
	Default     string // the default value
	Description string
	Verify      func(v string) error // optional verification method
}

func verifyInt(min, max int) func(v string) error {
	return func(v string) error {
		_, err := strconv.Atoi(v)
		return err
	}
}

func verifyEnum(list string) func(v string) error {
	return func(v string) error {
		items := strings.Split(list, "|")
		for _, i := range items {
			if i == v {
				return nil
			}
		}
		return fmt.Errorf("'%s' not in %s", v, list)
	}
}

func verifyBool(v string) error {
	if strings.ToLower(v) == "true" || strings.ToLower(v) == "false" {
		return nil
	}
	return fmt.Errorf("invalid boolean '%s'", v)
}

func DefaultConfiguration() Config {
	return Config{
		"app": ConfigGroup{
			"home": Setting{
				Description: "Path to the directory where factom related data such as the database is stored",
			},
		},
		"api": ConfigGroup{
			"port": Setting{
				Default:     "8088",
				Description: "The port of the Factomd API",
				Verify:      verifyInt(1, 65535),
			},
		},
		"network": ConfigGroup{
			"network": Setting{
				Default:     "MAIN",
				Description: "The network to connect to",
				Verify:      verifyEnum("MAIN|LOCAL|TEST|CUSTOM"),
			},
			"tlsenabled": Setting{
				Default:     "false",
				Description: "Enable TLS for the Control Panel and API",
				Verify:      verifyBool,
			},
		},
	}
}
