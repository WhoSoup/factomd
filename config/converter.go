package config

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var networkRegEx = regexp.MustCompile("(?i)^[a-z0-9_]+$") // also checks for blank
var timeRegEx = regexp.MustCompile("^(\\d+)(s|m|h|d)?$")
var hex64RegEx = regexp.MustCompile("^[a-fA-F0-9]{64}$")
var alphaRegEx = regexp.MustCompile("^[a-zA-Z0-9]*^")
var portRegEx = regexp.MustCompile("^[0-9]+$")
var urlRegEx = regexp.MustCompile("(?i)^(https?)://[^\\s/$.?#].[^\\s]*$")

func fNetwork(s string) bool {
	return networkRegEx.MatchString(s)
}

func fTime(s string) (int, error) {
	t := timeRegEx.FindStringSubmatch(s)
	if t == nil {
		return 0, fmt.Errorf("input must be a number followed by an optional 's' 'm' 'h' or 'd'")
	}

	v, err := strconv.Atoi(t[1])
	if err != nil {
		return 0, fmt.Errorf("unable to convert %s to an integer: %v", t[1], err)
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

func stringFTag(f, val string) error {
	switch f {
	case "url":
		if val == "" || urlRegEx.MatchString(val) {
			return nil
		}
		return fmt.Errorf("not a valid url")
	case "hex64":
		if hex64RegEx.MatchString(val) {
			return nil
		}
		return fmt.Errorf("input %s is not a valid hexademical string with 64 characters", val)
	case "network":
		if fNetwork(val) {
			return nil
		}
		return fmt.Errorf("network name contains invalid characters. use alphanumeric and _ (underscore) only")
	case "alpha":
		if alphaRegEx.MatchString(val) {
			return nil
		}
		return fmt.Errorf("setting contains non-alphanumeric characters")
	case "ipport":
		host, port, err := net.SplitHostPort(val) // local err
		if err != nil {
			return err
		}
		if len(host) < 1 {
			return fmt.Errorf("missing hostname in address \"%s\"", val)
		}
		if !portRegEx.MatchString(port) {
			return fmt.Errorf("missing port in address \"%s\"", val)
		}
		return nil
	default: // this is a developer error
		return fmt.Errorf("no string handler for f-tag \"%s\"", f)
	}
}

func intFTag(f, val string) (int, error) {
	switch f {
	case "time":
		time, err := fTime(val)
		if err != nil {
			return 0, err
		}
		return time, nil
	default: // this is a developer error
		return 0, fmt.Errorf("no int handler for f-tag \"%s\"", f)
	}
}

// Defines the kinds of variable types supported
func set(target reflect.Value, val string, tag reflect.StructTag) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unable to convert \"%s\" to %s: %v", val, target.Kind(), r)
		}
	}()

	val = strings.TrimSpace(val)

	switch target.Kind() {
	case reflect.Int:
		if f, ok := tag.Lookup("f"); ok {
			i, err := intFTag(f, val)
			if err != nil {
				return err // local err
			}
			target.SetInt(int64(i))
		} else {
			v, err := strconv.Atoi(val)
			if err != nil {
				return err // local err
			}
			target.SetInt(int64(v))
		}
	case reflect.Bool:
		target.SetBool(strings.ToLower(val) == "true" || strings.ToLower(val) == "yes")
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
			f, hasf := tag.Lookup("f")
			for k, i := range items {
				i = strings.TrimSpace(i)
				if hasf {
					if err = stringFTag(f, i); err != nil {
						return
					}
				}
				items[k] = i
			}
			target.SetString(strings.Join(items, ","))
		} else if f, ok := tag.Lookup("f"); ok {
			if err = stringFTag(f, val); err != nil {
				target.SetString(val)
			}
		} else {
			target.SetString(val)
		}
	default:
		err = fmt.Errorf("variable type \"%s\" does not have a handler in config/convert.go", target.Kind())
	}
	return
}
