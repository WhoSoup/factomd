package config

import (
	"fmt"
	"reflect"
	"strings"
)

// convert from golang's pascal case to camelcase used in config file
func lcFirst(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}
	s = strings.Replace(s, "P2P", "p2p", 1)
	s = strings.Replace(s, "API", "api", 1)
	s = strings.Replace(s, "DB", "db", 1)
	return strings.ToLower(string(s[0])) + string(s[1:])
}

// format enum list for display
func prettyEnum(s string) string {
	return strings.Replace(s, ",", ", ", -1)
}

func header() string {
	// TODO move version and build to globals or somewhere else
	return fmt.Sprintf("%7s Copyright (c) 2019 Factom Foundation\n", "Factomd")
}

func GetUsage() string {
	r := header()
	r += "Usage:\n"
	r += " All command line options supersede config file options.\n\n"
	c := new(Config)
	r += " -help\n -h -?\n"
	r += WordWrap("Prints this usage", 80, "    ") + "\n\n"
	r += fmt.Sprintf(" -%s %s\n", "config", "string")
	r += fmt.Sprintf(" -%s\n", "c")
	r += WordWrap("The path to the configuration file. Uses default location if left blank", 80, "    ") + "\n\n"
	err := c.walk(func(cat reflect.StructField, field reflect.StructField, val reflect.Value) error {
		if cat.Name != "Factomd" {
			return nil
		}
		var t string
		var enum string
		if tag, ok := field.Tag.Lookup("enum"); ok {
			t = " (enum)"
			enum = fmt.Sprintf("    Choices: %s\n", prettyEnum(tag))
		}

		t = val.Kind().String()
		if f, ok := field.Tag.Lookup("f"); ok {
			t = f
		}
		r += fmt.Sprintf(" -%s %s\n", lcFirst(field.Name), t)
		if short, ok := field.Tag.Lookup("short"); ok {
			r += fmt.Sprintf(" -%s\n", short)
		}
		if hint, ok := field.Tag.Lookup("hint"); ok {
			r += WordWrap(hint, 80, "    ") + "\n"
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

// WordWrap takes a multi-line string and word wraps each line according to the given character limit.
// Does not hyphenate words. The specified prefix will be prepended to every line.
// Lines that start with spaces will maintain their spaces for each wrapped line.
func WordWrap(s string, limit int, prefix string) string {
	lines := strings.Split(s, "\n")
	r := ""
	for i, line := range lines {
		trim := strings.TrimLeft(line, " ")
		diff := len(line) - len(trim)
		formatted := wrap(trim, limit-len(prefix)-diff)
		t := strings.Repeat(" ", diff)
		r += prefix + t + strings.Join(formatted, prefix+t)
		if i < len(lines)-1 {
			r += "\n"
		}
	}
	return r
}

func wrap(s string, limit int) []string {
	if len(s) <= limit {
		return []string{s}
	}
	tokens := strings.Split(s, " ")
	var r []string
	var line string
	for i, word := range tokens {
		if i > 0 && len(line)+len(word) > limit {
			line = strings.TrimRight(line, " ")
			r = append(r, line+"\n")
			line = ""
		}
		line += word
		if i < len(tokens)-1 {
			line += " "
		}
	}
	r = append(r, line)
	return r
}
