package config

import (
	"fmt"
	"os"
	"strings"
)

// Flags Parser
//
// There weren't any available flag parsing packages that suited this packages needs,
// so this is a very lightweight implementation.
//
// Scans through the slice for flags. Elements that start with "-" are considered "flags"
// and all other elements are considered "values".
//
// The flags have their dashes left-trimmed and the remainder is taken as name.
// If a value follows a flag, that value is associated with the flag.
// If a flag has no value, it is considered a boolean and will receive the value "true".
// Empty flag names and two values in a row are not permitted.
//
// Does not do any type or error checking for values, which is delegated to the rest of the
// package.
type Flags struct {
	flags   map[string]string
	checked map[string]bool
}

// ParseOSFlags returns a flag parser over the OS args
//
// golang handles arguments encapsulated in quotations and stores them in a single element
func ParseOSFlags() (*Flags, error) {
	return ParseFlags(os.Args[1:])
}

// ParseFlags returns a flag parser over an arbitrary slice of args
func ParseFlags(args []string) (*Flags, error) {
	f := new(Flags)
	f.flags = make(map[string]string)
	f.checked = make(map[string]bool)

	if err := f.parse(args); err != nil {
		return nil, err
	}
	return f, nil
}

// Get the value for a flag and mark it as used
func (f *Flags) Get(flag string) (string, bool) {
	flag = strings.ToLower(flag)
	val, ok := f.flags[flag]
	f.checked[flag] = true
	return val, ok
}

// GetS gets the value of a set of flags, marking them all as used, in the order of
// increasing priority. ([foo, f] will take the value of "f" over the value of "foo")
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

// Unused returns all of the flag names that have not been checked via Get or GetS. Used
// to obtain the list of unknown flags after error checking is done
func (f *Flags) Unused() []string {
	var r []string
	for k := range f.flags {
		if !f.checked[k] {
			r = append(r, k)
		}
	}
	return r
}

// process the slice
func (f *Flags) parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	el, args := args[0], args[1:]
	if len(el) > 0 && el[0] == '-' {
		var command, data string
		el = strings.TrimLeft(el, "-") // accept any number of prefix dashes

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

// add a case insensitive flag and value. returns false if the flag was already set
func (f *Flags) add(command, data string) bool {
	command = strings.ToLower(command)
	if _, ok := f.flags[command]; ok {
		return false
	}
	f.flags[command] = data
	return true
}
