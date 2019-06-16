package config

import "fmt"

type SettingType uint

const (
	INT SettingType = iota
	STRING
	BOOL
)

type Setting struct {
	Name    string
	Short   string
	Default string
	val1    string
	val2    int
	val3    bool
	Type    SettingType
	Verify  func(val string) error
}

func (s Setting) GetString() string {
	if s.Type != STRING {
		panic(fmt.Sprintf("setting %s is not a string", s.Name))
	}
	return s.val1
}

func (s Setting) GetInt() int {
	if s.Type != INT {
		panic(fmt.Sprintf("setting %s is not an integer", s.Name))
	}
	return s.val2
}

func (s Setting) GetBool() bool {
	if s.Type != BOOL {
		panic(fmt.Sprintf("setting %s is not a boolean", s.Name))
	}
	return s.val3
}
