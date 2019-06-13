package config

import (
	"fmt"
	"os"
	"strings"
)

// Flags There weren't any available flags libraries that did what I needed, so I made this
type Flags struct {
	flags   map[string]string
	checked map[string]bool
}

func ParseOSFlags() (*Flags, error) {
	return ParseFlags(os.Args[1:])
}

func ParseFlags(args []string) (*Flags, error) {
	f := new(Flags)
	f.flags = make(map[string]string)
	f.checked = make(map[string]bool)

	if err := f.parse(args); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Flags) Get(flag string) (string, bool) {
	flag = strings.ToLower(flag)
	val, ok := f.flags[flag]
	f.checked[flag] = true
	return val, ok
}

func (f *Flags) GetS(flags ...string) (string, bool) {
	var val string
	var ok bool

	for _, s := range flags {
		if v, k := f.Get(s); k {
			val = v
			ok = k
		}
	}
	return val, ok
}

func (f *Flags) Unused() []string {
	var r []string
	for k := range f.flags {
		if !f.checked[k] {
			r = append(r, k)
		}
	}
	return r
}

func (f *Flags) parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	el, args := args[0], args[1:]
	if len(el) > 0 && el[0] == '-' {
		var command, data string
		el = strings.TrimLeft(el, "-")

		if len(el) == 0 {
			return fmt.Errorf("empty flag name found")
		}

		if strings.Contains(el, "=") {
			split := strings.SplitN(el, "=", 2)
			command = split[0]
			data = split[1]
		} else if len(args) > 0 && args[0][0] != '-' {
			next := args[0]
			args = args[1:]
			command = el
			data = next
		} else {
			command = el
			data = "true"
		}

		if !f.add(command, data) {
			return fmt.Errorf("flag \"%s\" specified multiple times", el)
		}
	} else {
		return fmt.Errorf("unencapsulated data found: \"%s\"", el)
	}
	return f.parse(args)
}

func (f *Flags) add(command, data string) bool {
	command = strings.ToLower(command)
	if _, ok := f.flags[command]; ok {
		return false
	}
	f.flags[command] = data
	return true
}
