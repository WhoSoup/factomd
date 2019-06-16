package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/FactomProject/factomd/engine"
)

func lcFirst(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}
	s = strings.Replace(s, "P2P", "p2p", 1)
	return strings.ToLower(string(s[0])) + string(s[1:])
}

func prettyEnum(s string) string {
	return strings.Replace(s, ",", ", ", -1)
}

func GetUsage() string {
	b := engine.Build
	if b == "" {
		b = "dev"
	}
	r := fmt.Sprintf("////// Factomd v%s Build %s\n", engine.FactomdVersion, b)
	r += "Usage:\n"
	r += " All command line options supersede config file options."
	r += "\t-option int\n"
	r += "\t\tThis is some longer hint"

	var c Config
	err := walk(&c, func(cat reflect.StructField, field reflect.StructField, val reflect.Value) error {
		if cat.Name != "Factomd" {
			return nil
		}
		var t string
		var enum string
		if tag, ok := field.Tag.Lookup("enum"); ok {
			t = " (enum)"
			enum = fmt.Sprintf("    Choices: %s\n", prettyEnum(tag))
		}

		if f, ok := field.Tag.Lookup("f"); ok {
			t = " (" + f + ")"
		}
		r += fmt.Sprintf(" -%s %s%s\n", lcFirst(field.Name), val.Kind(), t)
		if hint, ok := field.Tag.Lookup("hint"); ok {
			r += fmt.Sprintf("    %s\n", hint)
		}
		r += enum
		r += "\n"
		return nil
	})

	if err != nil {
		panic(err)
	}

	return r
}
