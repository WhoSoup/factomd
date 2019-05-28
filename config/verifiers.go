package config

import (
	"fmt"
	"strconv"
	"strings"
)

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
