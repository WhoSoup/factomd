package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var networkRegEx = regexp.MustCompile("^[a-zA-Z0-9_]+$") // also checks for blank

func fNetwork(s string) bool {
	return networkRegEx.MatchString(s)
}

var timeRegEx = regexp.MustCompile("^(\\d+)(s|m|h|d)?$")

func fTime(s string) (int, error) {
	t := timeRegEx.FindStringSubmatch(s)
	if t == nil {
		return 0, fmt.Errorf("Input must be a number followed by an optional 's' 'm' 'h' or 'd'")
	}

	v, err := strconv.Atoi(t[1])
	if err != nil {
		return 0, fmt.Errorf("Unable to convert %s to an integer: %v", t[1], err)
	}

	switch t[2] {
	case "d": // days
		return v * 86400, nil
	case "h": // hours
		return v * 3600, nil
	case "m": // minutes
		return v * 60, nil
	case "s": // seconds
		fallthrough
	default:
		return v, nil
	}
}

func fEnum(s string, set string) (string, bool) {
	val := strings.Split(strings.ToUpper(set), ",")
	s = strings.ToUpper(strings.TrimSpace(s))
	for _, v := range val {
		if s == strings.TrimSpace(v) {
			return s, true
		}
	}
	return "", false
}

// Defines the kinds of variable types supported
func set(target reflect.Value, val string, tag reflect.StructTag) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Unable to convert \"%s\" to %s", val, target.Kind())
		}
	}()

	val = strings.TrimSpace(val)

	switch target.Kind() {
	case reflect.Int:
		v, err := strconv.Atoi(val)
		if err != nil {
			return err // local err
		}
		target.SetInt(int64(v))
	case reflect.Bool:
		target.SetBool(strings.ToLower(val) == "true")
	case reflect.String:
		//
		// ENUM
		//
		if set, ok := tag.Lookup("enum"); ok {
			if choice, ok := fEnum(val, set); ok {
				target.SetString(choice)
			} else {
				err = fmt.Errorf("%s not part of enum %s", val, set)
			}
		} else if sep, ok := tag.Lookup("list"); ok {
			items := strings.Split(val, sep)
			for _, i := range items {
				i = strings.TrimSpace(i)

			}
		} else if f, ok := tag.Lookup("f"); ok {
			switch f {
			case "time":

			default:
				panic(fmt.Sprintf("could not find string handler %s", f))
			}
		} else {
			target.SetString(val)
		}
	default:
		err = fmt.Errorf("variable type \"%s\" does not have a handler in config/convert.go", target.Kind())
	}
	return
}

func fString(val string, f string) (string, bool) {
	switch f {

	default:
		panic(fmt.Sprintf("f-tag method function handler for %s not found", f))
	}
}
